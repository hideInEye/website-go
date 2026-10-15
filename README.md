# 日落的足迹 · Go 后端

直接基于 [go-admin-template](https://github.com/hideInEye/go-admin-template) 迁移博客业务，保留模板 Git 历史、启动入口、Dig 注入和 `ctl → bll → model` 分层，不另建聚合应用包。启动链为 `services/main.go → app.Init → BuildContainer / InitStore → web.RegisterRouter`，请求继续交给控制器、业务接口及实现、模型接口及 Gorm 实现。原示例路由和示例表自动迁移不运行。

新仓库：`git@github.com:hideInEye/website-go.git`，分支 `codex/blog-api`。**只推送 origin，不向 template 推送**。原 Nest 后端 `website-server` 保持不动；现有数据库、管理员、密码哈希和文章继续使用。

> 以下为该分支的运行方式；单元测试和编译不代表数据库、MinIO、邮件或线上部署已经完成联调，不要用早期 WIP 检查点替换线上服务。

## 本地开发：前后端分别启动

本地需要 Go 1.26，以及已有 MySQL、MinIO。两个项目不必放在同一目录，不用 Docker 启动开发环境。

```bash
git clone --branch codex/blog-api git@github.com:hideInEye/website-go.git website-go-server
cd website-go-server
chmod 600 config/config.toml
```

只编辑 `config/config.toml`，在 `[mysql]` 中填写 `host/port/db_name/user/password`，在 `[jwt_auth]` 中配置模板 JWT，在 `[minio]` 中填写 `addr/assess_key/secret_key/bucket`。`assess_key` 沿用模板原字段拼写，不要改成 `access_key`。本地 `[site].origin = "http://localhost:3000"`，MySQL 使用 `localhost`，MinIO 使用 `addr = "localhost:9000"`，API 的 `[http].port = 4000`。

请从仓库根目录运行命令。本地和服务器都只读取 `config/config.toml`，分别填写各自机器的参数，不需要复制其他环境配置文件。后端不加载根目录 `.env`，也不使用 `DATABASE_URL` 或 `MYSQL_*` 等环境变量覆盖业务配置。该 TOML 受 Git 跟踪，填写真实密码后不要提交这些敏感值。模板原有 `config/model.conf` 是 Casbin 权限模型规则，不是另一份环境配置，保持不动。

```bash
go mod download
# 首次启动前，建表或校验已有表，不导入或覆盖数据
go run services/main.go migrate -c config/config.toml
go run services/main.go -c config/config.toml
```

JWT 使用模板的 `InitAuth → pkg/auth/jwtauth`，默认 `HS512`、有效期 `7200` 秒，也支持模板原有的 `HS256/HS384`。不再额外限制密钥至少 32 字符；生产环境应使用随机私密密钥，不要沿用模板公开的示例密钥。`store = "file"` 时令牌撤销记录保存在 `file_path` 指定的位置，默认 `data/jwt_auth.db`；本地启动会创建该目录。

API 启动后，可在另一个终端手动执行 `go run services/main.go healthcheck -c config/config.toml` 检查存活状态。该命令只访问本机 API，不创建测试数据库或用户；数据库就绪情况由 `/api/v1/health/ready` 返回。

数据库必须已存在。新空库需先由数据库管理员创建 `utf8mb4 / utf8mb4_unicode_ci` 数据库并授权应用账号；已有博客库无需再次建库或创建管理员。初始化 SQL 沿用原 Prisma migration，包含 MySQL ngram 全文索引，不支持用 SQLite 替代。

另开终端，在前端目录运行：

```bash
cp .env.example .env.local
pnpm install --frozen-lockfile
pnpm dev
```

访问 `http://localhost:3000`。前端仍按 Vite 方式使用自己的 `.env.local`，不与后端配置混用。Vite 将 `/api/v1` 转发到 API，`/media/` 转发到 MinIO。后端 `[media].upload_public_url`、`public_url` 留空，使用同源代理；前端 `MEDIA_PROXY_HOST` 必须与后端签名时的 `[minio].addr` 一致，本地默认 `localhost:9000`。

## GitHub 构建镜像

推送本仓库的 `codex/blog-api`、`main` 或 `master`，工作流 `Publish Go API image` 在 GitHub 上构建并发布：

```text
ghcr.io/hideineye/sunset-footprints-api-go:deploy
ghcr.io/hideineye/sunset-footprints-api-go:sha-<完整提交SHA>
```

前端 `codex/vite-react-go-rebuild` 分支发布：

```text
ghcr.io/hideineye/sunset-footprints-web:vite
ghcr.io/hideineye/sunset-footprints-web:sha-<完整提交SHA>
```

新镜像不覆盖旧 Next / Nest 的 `:deploy` 镜像。API 最终镜像只包含 Go 二进制、证书和时区文件，不打包 TOML 配置或密码；运行时将宿主机 `config/config.toml` 只读挂载到 `/app/config/config.toml`。Web 最终容器只有 Nginx 与静态资源。服务器不需要 Node.js 或 Go，不执行 `docker build`。

## 服务器逐步部署

前提：Docker、Compose v2 和宿主机 MySQL、MinIO、Nginx 已安装，两个新镜像都已发布成功。服务器只需要新后端仓库，Compose 同时管理 Web/API。

### 1. 拉代码并准备配置文件

```bash
git clone --branch codex/blog-api git@github.com:hideInEye/website-go.git website-go-server
cd website-go-server
chmod 600 config/config.toml
```

用部署账号编辑 `config/config.toml`，脚本会按该文件所有者自动设置 API 容器 UID/GID，并准备 `data/` 目录。使用同一账号运行脚本即可，不需要另外填写运行用户配置。如果已有目录所有者不同，先明确调整所有权；脚本不会自动执行 sudo 或 chown。Compose 将 `data/` 挂载到 `/app/data`，容器重建后保留模板 JWT 文件存储。

在唯一的 `config/config.toml` 对应分组中修改以下字段，不要把示例再追加成重复分组；其他模板选项保持原值：

```toml
run_mode = "release"

[http]
host = "0.0.0.0"
port = 4000
trusted_proxies = ["127.0.0.1/32", "::1/128", "172.16.0.0/12"]

[site]
origin = "http://admin.zylife.top"
allow_insecure_http = true

[mysql]
host = "host.docker.internal"
port = 3336
db_name = "sunset_blog"
user = "sunset"
password = "replace-with-existing-mysql-password"

[jwt_auth]
signing_method = "HS512"
signing_key = "replace-with-your-private-jwt-signing-key"
expired = 7200
store = "file"
file_path = "data/jwt_auth.db"
refresh_ttl_days = 30
cookie_secure = false

[minio]
addr = "host.docker.internal:9000"
assess_key = "replace-with-existing-minio-access-key"
secret_key = "replace-with-existing-minio-secret-key"
bucket = "sunset-media"
use_ssl = false
request_timeout_ms = 5000

[mail]
mode = "log"

[media]
upload_public_url = ""
public_url = ""
```

`[mysql].port` 改成宿主机真实端口。数据库填已经迁移后的博客库，不要误填结构不同的旧 `web_site_server.posts`。已有账号、密码和管理员权限不变，不自动建立新管理员。JWT 改用模板的 `sub` 用户标识，原先自定义 `id` 字段的访问令牌不再有效，切换后重新登录即可；现有数据库刷新会话仍可换取新令牌。保留示例的 `[mysql].parameters`，日期解析和字符集设置沿用模板。

`docker-compose.yml` 默认把 Web/API 映射到 `127.0.0.1:18080/18081`，不会占用宿主机 80。前端代理默认 `MEDIA_UPSTREAM=http://host.docker.internal:9000`、`MEDIA_PROXY_HOST=host.docker.internal:9000`，与 TOML 的 `[minio].addr` 对应；API 容器内 `[http].port` 保持 4000。默认值无需修改，端口或镜像的临时覆盖可直接传给部署命令，不需要额外配置文件。

容器中 `localhost` 指容器自己，因此连接宿主机用 `host.docker.internal`，Compose 已配置 host-gateway。MySQL/MinIO 不需要开放公网，但需允许 Docker 网关访问。`[http].trusted_proxies` 示例信任 loopback 和 Docker 常用 `172.16.0.0/12` 网段，若使用自定义网段请收窄/改为实际反向代理网段，不要设成 `0.0.0.0/0`。

若 GHCR 镜像为私有，先 `docker login ghcr.io`，使用具备 `read:packages` 的凭据；或将对应镜像包设为公开。不要将令牌写进 Git。

### 2. 准备切换端口

新项目名 `sunset-footprints-go` 不会删除旧 `sunset-footprints` 容器，但默认仍使用 18080/18081。正式切换前先备份，然后在**旧后端目录**停止旧 Web/API：

```bash
docker compose --env-file .env stop web api
```

以上 `.env` 仅属于旧 Nest 项目。这不会停止宿主机 MySQL/MinIO，也不删数据库。若想保留旧服务同时检查新服务，使用 `HTTP_PORT=28080 API_PORT=28081 sh scripts/deploy.sh`，并同步调整站点入口/反向代理。

### 3. 拉镜像并启动

回到新 `website-go-server` 目录：

```bash
sh scripts/preflight.sh
sh scripts/deploy.sh
docker compose ps
docker compose logs --tail=100 api web
```

脚本不读取额外 `.env` 文件。预检核对 `config/config.toml` 存在、`400/600` 权限、容器 UID/GID、JWT 数据目录权限和 Compose 格式，不连接数据库或运行应用，也不声称已验证 TOML 业务值。

部署依次拉 API/Web 镜像，运行一次轻量 `migrate`，然后启动 API/Web。API 与 migrate 读取同一个只读挂载的 TOML，不接收数据库或 JWT 等环境变量。`migrate` 调用模板 Gorm 模型目录内的初始化函数，对空库建表，对已有表核对字段/索引名称并补齐缺失外键，不导入、重建或清空已有数据。初始化 SQL 位于 `internal/app/model/impl/gorm/schema.sql`，与原 Prisma migration 一致。名称校验不是完整的列类型审计；不要自行改已有表后假定迁移会自动修复。MySQL DDL 不保证事务性，中途失败时先看日志，修复后重跑，不要删除表。

### 4. 宿主机 Nginx

参考 [deploy/nginx-host.conf.example](deploy/nginx-host.conf.example)，替换域名后配置到宿主机。宿主机监听 80，将整个站点转发到 `127.0.0.1:18080`；Web 容器内部监听 3000，API 映射到 `127.0.0.1:18081`，不会与宿主机 80 冲突。

```bash
sudo nginx -t
sudo systemctl reload nginx
```

Web 容器负责 `/api/`、`/media/`、RSS、sitemap、robots 和旧文章链接，其余地址 SPA 回退，刷新深层路由仍可打开。若宿主机已有单独 `/api/` 或 `/media/` 规则，检查它们是否还指向旧项目。

HTTP 使用 TOML 中 `[jwt_auth].cookie_secure = false`、`[site].allow_insecure_http = true`，会明文传输登录密码/Cookie。以后用 HTTPS 时，在宿主机终止 TLS，设置 `[site].origin = "https://..."`、`[jwt_auth].cookie_secure = true`；MinIO 内部仍可使用 HTTP。

## MinIO 与邮箱

继续使用已经部署的桶和访问策略。浏览器上传地址由 `[site].origin` 派生为站点下的 `/media/<bucket>/<object>`，Nginx 移除 `/media` 并恢复签名 Host，不能自行修改签名查询参数。后端不自动改变桶的公开访问策略；图片匿名读取权限按原部署保留。

默认 `[mail].mode = "log"`：原注册、验证、找回密码流程保留，链接写入 API 日志，不实际发信。日志包含敏感令牌，不要公开。启用 SMTP 时在同一 TOML 文件设置 `[mail].mode = "smtp"`，填写该分组的 `host/port/user/password/from`。本次开发不会连接真实 SMTP。

## 更新、备份与回滚

更新：等 GitHub 镜像成功，先在仓库外备份服务器填写的 `config/config.toml`，再拉取代码并确认本机参数未被覆盖，然后重新执行 `sh scripts/deploy.sh`。不要为了拉代码丢弃含服务器配置的本地改动。锁定或回滚镜像时，用 `WEB_IMAGE=目标镜像 API_IMAGE=目标镜像 sh scripts/deploy.sh` 指定 `sha-<完整提交SHA>` 镜像。修改业务 TOML 后重启 API，使进程重新读取配置。

正常部署不需要备份工具参数。可选备份脚本 `sh scripts/backup.sh ./backups` 使用调用终端临时导出的 `MYSQL_HOST/MYSQL_PORT/MYSQL_DATABASE/MYSQL_USER/MYSQL_PASSWORD`、`MINIO_ENDPOINT/MINIO_PORT/MINIO_USE_SSL/MINIO_ACCESS_KEY/MINIO_SECRET_KEY/MINIO_BUCKET`；不再创建备份配置文件。这些只传给备份客户端，不覆盖 TOML 或传给 API。凭据请用终端安全输入或密钥管理工具注入，不要将真实密码写进可保存的命令历史。数据库和桶填写实际要备份的目标，账号需具有相应备份权限。

备份脚本会先检查必需参数，再短暂停止此项目 Web/API，完成后恢复；不会新建数据库服务。含删库和覆盖对象的自动恢复脚本尚未迁移，需单独授权与核对恢复范围，不能作为部署默认步骤。

切回旧 Nest 项目：先停新项目 Web/API，再在旧目录启动原 Web/API，避免端口冲突；本次数据库结构沿用旧博客，不需要反向数据迁移。

## 代码与验证

按模板原目录迁移业务：

- `services/main.go`：命令入口，支持默认启动、`migrate` 和 `healthcheck`，均可用 `-c` 指定 TOML。
- `internal/app/app.go`、`store.go`、`web.go`：模板初始化、依赖注入、存储及 HTTP 生命周期。
- `internal/app/routers/web/`、`ctl/`：路由、请求解析、响应；不直接操作数据库。
- `internal/app/bll/`、`bll/impl/internal/`：业务接口与实现，处理权限、发布、审核、认证等规则。
- `internal/app/model/`、`model/impl/gorm/internal/model/`：存储接口与 Gorm 实现；`entity/` 映射现有 camelCase 数据库列。
- `internal/app/schema/`：按业务拆分的 DTO、查询参数和响应类型；`internal/app/config/` 沿用模板 TOML 配置入口，配置文件统一位于 `config/`。
- `internal/app/model/impl/gorm/migrate.go`、`schema.sql`：仅显式运行 `migrate` 时检查和初始化博客表；不使用示例 AutoMigrate。

GitHub 的 `Go API checks` 对 `internal/app`、`services`、`pkg/minio` 检查官方 `gofmt`，执行无需生产服务的 Go 检查并编译 `./services`。`Publish Go API image` 负责生成部署镜像。

本地可运行 `go test ./...` 和 `go build -o /tmp/website-go-api services/main.go`；与镜像一致的构建参数为 `CGO_ENABLED=0 go build -tags nomsgpack -o /tmp/website-go-api ./services`。JWT 回归测试使用临时文件存储与业务桩，不连接实际 MySQL、MinIO 或 SMTP。数据库行为、浏览器视觉、容器代理及镜像发布仍需单独验收，不应把单元测试和构建通过视为线上部署已验证。
