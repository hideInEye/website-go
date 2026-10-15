package blog

import (
 "fmt"
 "log"
 "net/http"
 "strings"
 "time"
 "go-gin/pkg/gormplus"
 "github.com/gin-gonic/gin"
 "github.com/google/uuid"
 "github.com/jinzhu/gorm"
 mysql "github.com/go-sql-driver/mysql"
 minio "github.com/minio/minio-go/v7"
 "github.com/minio/minio-go/v7/pkg/credentials"
)

type App struct { DB *gorm.DB; Config Config; Storage *minio.Client }

func New(c Config) (*App,error) {
 dsn,err:=c.DSN();if err!=nil{return nil,err}
 db,err:=gormplus.New(&gormplus.Config{DBType:"mysql",DSN:dsn,MaxOpenConns:8,MaxIdleConns:2,MaxLifetime:180});if err!=nil{return nil,err}
 // Reuse the template's Gorm adapter without its demo-table AutoMigrate.
 transport:=http.DefaultTransport.(*http.Transport).Clone();transport.ResponseHeaderTimeout=c.MinioTimeout
 storage,err:=minio.New(fmt.Sprintf("%s:%d",c.MinioEndpoint,c.MinioPort),&minio.Options{Creds:credentials.NewStaticV4(c.MinioAccessKey,c.MinioSecretKey,""),Secure:c.MinioSSL,Transport:transport,Region:"us-east-1"});if err!=nil{db.Close();return nil,err}
 return &App{DB:db.DB,Config:c,Storage:storage},nil
}

func (a *App) Router() *gin.Engine {
 if a.Config.Environment=="production"{gin.SetMode(gin.ReleaseMode)}
 r:=gin.New();r.Use(gin.Logger(),gin.Recovery(),a.headers())
 r.SetTrustedProxies([]string{"127.0.0.1","::1"})
 r.NoRoute(func(c *gin.Context){fail(c,404,"接口不存在")})
 api:=r.Group("/api/v1")
 api.GET("/health/live",func(c *gin.Context){respond(c,200,gin.H{"status":"ok"})})
 api.GET("/health/ready",func(c *gin.Context){if err:=a.DB.DB().PingContext(c.Request.Context());err!=nil{fail(c,503,"数据库暂时不可用");return};respond(c,200,gin.H{"status":"ok","info":gin.H{"database":gin.H{"status":"up"}},"error":gin.H{},"details":gin.H{"database":gin.H{"status":"up"}}})})
 a.registerAuth(api);a.registerContent(api);a.registerMedia(api);a.registerFeeds(r)
 return r
}

func (a *App) headers() gin.HandlerFunc {return func(c *gin.Context){
 c.Header("X-Content-Type-Options","nosniff");c.Header("X-Frame-Options","DENY");c.Header("Referrer-Policy","strict-origin-when-cross-origin");c.Header("X-Request-ID",newID())
 origin:=c.GetHeader("Origin");if origin==a.Config.AppOrigin{c.Header("Access-Control-Allow-Origin",origin);c.Header("Vary","Origin");c.Header("Access-Control-Allow-Credentials","true");c.Header("Access-Control-Allow-Headers","Content-Type, Authorization");c.Header("Access-Control-Allow-Methods","GET, POST, PUT, PATCH, DELETE, OPTIONS")}
 if c.Request.Method==http.MethodOptions{if origin!=""&&origin!=a.Config.AppOrigin{fail(c,403,"请求来源未被允许");return};c.AbortWithStatus(204);return}
 if strings.Contains("|POST|PUT|PATCH|DELETE|","|"+c.Request.Method+"|")&&origin!=""&&origin!=a.Config.AppOrigin{fail(c,403,"请求来源未被允许");return}
 c.Request.Body=http.MaxBytesReader(c.Writer,c.Request.Body,12*1024*1024);c.Next()
}}

func newID() string {return strings.ReplaceAll(uuid.NewString(),"-","")}
func respond(c *gin.Context,status int,data interface{}) {c.JSON(status,gin.H{"data":data})}
func fail(c *gin.Context,status int,message string) {c.AbortWithStatusJSON(status,gin.H{"statusCode":status,"code":fmt.Sprintf("HTTP_%d",status),"message":message,"traceId":c.Writer.Header().Get("X-Request-ID")})}
func dbError(c *gin.Context,err error) bool {
 if err==nil{return false};if gorm.IsRecordNotFoundError(err){fail(c,404,"记录不存在");return true}
 if e,ok:=err.(*mysql.MySQLError);ok&&(e.Number==1062||e.Number==1451||e.Number==1452){fail(c,409,"数据重复或仍被其他记录使用");return true}
 log.Printf("database operation failed: %v",err);fail(c,500,"数据操作失败");return true
}
func bind(c *gin.Context,value interface{}) bool {if err:=c.ShouldBindJSON(value);err!=nil{fail(c,400,"请求参数格式或内容不正确");return false};return true}
func transaction(db *gorm.DB,fn func(*gorm.DB)error) error {tx:=db.Begin();if tx.Error!=nil{return tx.Error};defer func(){if p:=recover();p!=nil{tx.Rollback();panic(p)}}();if err:=fn(tx);err!=nil{tx.Rollback();return err};return tx.Commit().Error}
func utcNow() time.Time {return time.Now().UTC()}
