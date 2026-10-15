package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

var (
	global *Config
)

// LoadGlobalConfig 加载全局配置
func LoadGlobalConfig(fpath string) error {
	c, err := ParseConfig(fpath)
	if err != nil {
		return err
	}
	if err := c.Validate(); err != nil {
		return err
	}
	global = c
	return nil
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *Config {
	if global == nil {
		return &Config{}
	}
	return global
}

// ParseConfig 解析配置文件
func ParseConfig(fpath string) (*Config, error) {
	var c Config
	_, err := toml.DecodeFile(fpath, &c)
	if err != nil {
		return nil, err
	}
	c.setDefaults()
	return &c, nil
}

// Redis redis配置参数
type Redis struct {
	Enable   bool   `toml:"enable"`
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

// Config 配置参数
type Config struct {
	RunMode         string      `toml:"run_mode"`
	CasbinModelConf string      `toml:"casbin_model_conf"`
	WWW             string      `toml:"www"`
	Swagger         string      `toml:"swagger"`
	Store           string      `toml:"store"`
	AllowInitMenu   bool        `toml:"allow_init_menu"`
	EnableCasbin    bool        `toml:"enable_casbin"`
	Log             Log         `toml:"log"`
	LogGormHook     LogGormHook `toml:"log_gorm_hook"`
	Root            Root        `toml:"root"`
	JWTAuth         JWTAuth     `toml:"jwt_auth"`
	HTTP            HTTP        `toml:"http"`
	Monitor         Monitor     `toml:"monitor"`
	Captcha         Captcha     `toml:"captcha"`
	RateLimiter     RateLimiter `toml:"rate_limiter"`
	CORS            CORS        `toml:"cors"`
	Redis           Redis       `toml:"redis"`
	Gorm            Gorm        `toml:"gorm"`
	MySQL           MySQL       `toml:"mysql"`
	Postgres        Postgres    `toml:"postgres"`
	Sqlite3         Sqlite3     `toml:"sqlite3"`
	Upload          Upload      `toml:"upload"`
	Minio           Minio       `toml:"minio"`
	AMap            AMap        `toml:"amap"`
	Flow            Flow        `toml:"flow"`
	Push            Push        `toml:"push"`
	Auth            Auth        `toml:"auth"`
	Site            Site        `toml:"site"`
	Mail            Mail        `toml:"mail"`
	Media           Media       `toml:"media"`
}

// Check 检查桶是否存在
func (a FileUploadBuckets) Check(name string) bool {
	for _, item := range a {
		if item == name {
			return true
		}
	}
	return false
}

// Log 日志配置参数
type Log struct {
	Level         int    `toml:"level"`
	Format        string `toml:"format"`
	Output        string `toml:"output"`
	OutputFile    string `toml:"output_file"`
	EnableHook    bool   `toml:"enable_hook"`
	Hook          string `toml:"hook"`
	HookMaxThread int    `toml:"hook_max_thread"`
	HookMaxBuffer int    `toml:"hook_max_buffer"`
}

// LogGormHook 日志gorm钩子配置
type LogGormHook struct {
	DBType       string `toml:"db_type"`
	MaxLifetime  int    `toml:"max_lifetime"`
	MaxOpenConns int    `toml:"max_open_conns"`
	MaxIdleConns int    `toml:"max_idle_conns"`
	Table        string `toml:"table"`
}

// Root root用户
type Root struct {
	UserName string `toml:"user_name"`
	Password string `toml:"password"`
	RealName string `toml:"real_name"`
}

// JWTAuth 用户认证
type JWTAuth struct {
	RefreshTTLDays int    `toml:"refresh_ttl_days"`
	CookieSecure   bool   `toml:"cookie_secure"`
	SigningMethod  string `toml:"signing_method"`
	SigningKey     string `toml:"signing_key"`
	Expired        int    `toml:"expired"`
	Store          string `toml:"store"`
	FilePath       string `toml:"file_path"`
	RedisDB        int    `toml:"redis_db"`
	RedisPrefix    string `toml:"redis_prefix"`
	ThridKey       string `toml:"third_key"`
}

// HTTP http配置参数
type HTTP struct {
	TrustedProxies  []string `toml:"trusted_proxies"`
	Host            string   `toml:"host"`
	Port            int      `toml:"port"`
	ShutdownTimeout int      `toml:"shutdown_timeout"`
}

// Monitor 监控配置参数
type Monitor struct {
	Enable    bool   `toml:"enable"`
	Addr      string `toml:"addr"`
	ConfigDir string `toml:"config_dir"`
}

// Captcha 图形验证码配置参数
type Captcha struct {
	Store       string `toml:"store"`
	Length      int    `toml:"length"`
	Width       int    `toml:"width"`
	Height      int    `toml:"height"`
	RedisDB     int    `toml:"redis_db"`
	RedisPrefix string `toml:"redis_prefix"`
}

// RateLimiter 请求频率限制配置参数
type RateLimiter struct {
	Enable  bool  `toml:"enable"`
	Count   int64 `toml:"count"`
	RedisDB int   `toml:"redis_db"`
}

// CORS 跨域请求配置参数
type CORS struct {
	Enable           bool     `toml:"enable"`
	AllowOrigins     []string `toml:"allow_origins"`
	AllowMethods     []string `toml:"allow_methods"`
	AllowHeaders     []string `toml:"allow_headers"`
	AllowCredentials bool     `toml:"allow_credentials"`
	MaxAge           int      `toml:"max_age"`
}

// Gorm gorm配置参数
type Gorm struct {
	Debug        bool   `toml:"debug"`
	DBType       string `toml:"db_type"`
	MaxLifetime  int    `toml:"max_lifetime"`
	MaxOpenConns int    `toml:"max_open_conns"`
	MaxIdleConns int    `toml:"max_idle_conns"`
	TablePrefix  string `toml:"table_prefix"`
	AutoMigrate  bool   `toml:"auto_migrate"`
}

// MySQL mysql配置参数
type MySQL struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	DBName     string `toml:"db_name"`
	Parameters string `toml:"parameters"`
}

// DSN 数据库连接串
func (a MySQL) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		a.User, a.Password, a.Host, a.Port, a.DBName, a.Parameters)
}

// Postgres postgres配置参数
type Postgres struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"db_name"`
	SSLMode  string `toml:"sslmode"`
}

// DSN 数据库连接串
func (a Postgres) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		a.Host, a.Port, a.User, a.DBName, a.Password, a.SSLMode)
}

// Sqlite3 sqlite3配置参数
type Sqlite3 struct {
	Path string `toml:"path"`
}

// DSN 数据库连接串
func (a Sqlite3) DSN() string {
	return a.Path
}

// Upload 文件上传配置参数
type Upload struct {
	Prefix    string            `toml:"prefix"`
	SizeLimit int64             `toml:"size_limit"`
	Buckets   FileUploadBuckets `toml:"buckets"`
}

// FileUploadBuckets 文件上传桶列表
type FileUploadBuckets []string

// Minio minio配置参数
type Minio struct {
	Bucket           string `toml:"bucket"`
	UseSSL           bool   `toml:"use_ssl"`
	RequestTimeoutMS int    `toml:"request_timeout_ms"`
	Addr             string `toml:"addr"`
	AssessKey        string `toml:"assess_key"`
	SecretKey        string `toml:"secret_key"`
}

// Site 站点地址与 HTTP 部署配置。
type Site struct {
	Origin            string `toml:"origin"`
	AllowInsecureHTTP bool   `toml:"allow_insecure_http"`
}

// Mail 注册验证与密码重置邮件配置。
type Mail struct {
	Mode     string `toml:"mode"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	From     string `toml:"from"`
}

// Media 浏览器访问媒体所使用的公开代理地址。
type Media struct {
	PublicURL       string `toml:"public_url"`
	UploadPublicURL string `toml:"upload_public_url"`
}

// AMap 高德地图配置参数
type AMap struct {
	Key string `toml:"key"`
}

// ExcelExport excel导出配置参数
type ExcelExport struct {
	ProjCalculateExcel    string `toml:"proj_calculate_excel"`
	ProjectTpl            string `toml:"project_tpl"`
	MarketingArchive      string `toml:"marketing_archive"`
	ResumeTemplate        string `toml:"resume_template"`
	FundApplied           string `toml:"fund_applied"`
	FundTransferApply     string `toml:"fund_transfer_apply"`
	NoFundTransferApply   string `toml:"no_fund_transfer_apply"`
	NoContractApplied     string `toml:"no_contract_applied"`
	CostContractExport    string `toml:"cost_contract_export"`
	CostDesignExport      string `toml:"cost_design_export"`
	CostVisaExport        string `toml:"cost_visa_export"`
	CostMaterialsExport   string `toml:"cost_materials_export"`
	CostDynamicExport     string `toml:"cost_dynamic_export"`
	ProjectAdjustExcel    string `toml:"project_adjust_excel"`
	ContractPlanningExcel string `toml:"contract_planning_excel"`
	TenantImport          string `toml:"tenant_import"`
	ProjectExcel          string `toml:"project_excel"`
	MonthlyPlanExcel      string `toml:"monthly_plan_excel"`
}

type Flow struct {
	Addr       string `toml:"addr"`
	SystemCode string `toml:"system_code"`
}

type ExportPDF struct {
	FontsPath string `toml:"fonts_path"`
}

type Push struct {
	Addr string `toml:"addr"`
}

// Auth 认证中心配置参数
type Auth struct {
	Addr         string `toml:"addr"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	Identify     string `toml:"identify"`
}
