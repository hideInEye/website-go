package blog

import (
 "fmt"
 "net/url"
 "os"
 "strconv"
 "strings"
 "time"
 "github.com/go-sql-driver/mysql"
 "github.com/joho/godotenv"
)

type Config struct {
 Environment string
 Port int
 AppOrigin string
 DatabaseURL string
 JWTSecret string
 AccessTTL time.Duration
 RefreshDays int
 CookieSecure bool
 MailMode string
 SMTPHost string
 SMTPPort int
 SMTPUser string
 SMTPPassword string
 MailFrom string
 MinioEndpoint string
 MinioPort int
 MinioSSL bool
 MinioAccessKey string
 MinioSecretKey string
 MinioBucket string
 MinioTimeout time.Duration
 MediaPublicURL string
 MediaUploadPublicURL string
}

func env(name, fallback string) string { if value := os.Getenv(name); value != "" { return value }; return fallback }
func LoadConfig() (Config, error) {
 // Existing process variables always win over local .env values.
 if err := godotenv.Load(); err != nil && !os.IsNotExist(err) { return Config{}, fmt.Errorf("load .env: %w", err) }
 c := Config{Environment:env("NODE_ENV","development"), AppOrigin:strings.TrimRight(env("APP_ORIGIN",env("SITE_URL","http://localhost:3000")),"/"), DatabaseURL:os.Getenv("DATABASE_URL"), JWTSecret:os.Getenv("JWT_ACCESS_SECRET"), MailMode:env("MAIL_MODE","log"), SMTPHost:env("SMTP_HOST","localhost"), SMTPUser:os.Getenv("SMTP_USER"), SMTPPassword:os.Getenv("SMTP_PASSWORD"), MailFrom:env("MAIL_FROM","日落的足迹 <noreply@example.com>"), MinioEndpoint:env("MINIO_ENDPOINT","localhost"), MinioAccessKey:os.Getenv("MINIO_ACCESS_KEY"), MinioSecretKey:os.Getenv("MINIO_SECRET_KEY"), MinioBucket:env("MINIO_BUCKET","sunset-media"), MediaPublicURL:strings.TrimRight(os.Getenv("MEDIA_PUBLIC_URL"),"/"), MediaUploadPublicURL:strings.TrimRight(os.Getenv("MEDIA_UPLOAD_PUBLIC_URL"),"/")}
 for _, field := range []struct{name, fallback string; target *int; min, max int}{{"PORT","4000",&c.Port,1,65535},{"SMTP_PORT","1025",&c.SMTPPort,1,65535},{"MINIO_PORT","9000",&c.MinioPort,1,65535},{"REFRESH_TOKEN_TTL_DAYS","30",&c.RefreshDays,1,3650}} { n,err:=strconv.Atoi(env(field.name,field.fallback)); if err!=nil||n<field.min||n>field.max{return c,fmt.Errorf("invalid %s",field.name)}; *field.target=n }
 var err error
 c.AccessTTL,err=time.ParseDuration(env("ACCESS_TOKEN_TTL","15m")); if err!=nil||c.AccessTTL<=0{return c,fmt.Errorf("invalid ACCESS_TOKEN_TTL")}
 timeout,err:=strconv.Atoi(env("MINIO_REQUEST_TIMEOUT_MS","5000"));if err!=nil||timeout<1||timeout>30000{return c,fmt.Errorf("invalid MINIO_REQUEST_TIMEOUT_MS")};c.MinioTimeout=time.Duration(timeout)*time.Millisecond
 c.CookieSecure,err=strconv.ParseBool(env("COOKIE_SECURE","false"));if err!=nil{return c,fmt.Errorf("invalid COOKIE_SECURE")}
 c.MinioSSL,err=strconv.ParseBool(env("MINIO_USE_SSL","false"));if err!=nil{return c,fmt.Errorf("invalid MINIO_USE_SSL")}
 u,err:=url.Parse(c.AppOrigin);if err!=nil||u.Host==""||(u.Scheme!="http"&&u.Scheme!="https")||u.Path!=""||u.RawQuery!=""||u.Fragment!=""||u.User!=nil{return c,fmt.Errorf("APP_ORIGIN must be an HTTP(S) origin")}
 if len(c.JWTSecret)<32{return c,fmt.Errorf("JWT_ACCESS_SECRET must contain at least 32 characters")}
 if _,err=c.DSN();err!=nil{return c,err}
 if c.MailMode!="log"&&c.MailMode!="smtp"{return c,fmt.Errorf("MAIL_MODE must be log or smtp")}
 if strings.ContainsAny(c.MinioEndpoint,"/:")||c.MinioEndpoint==""||c.MinioAccessKey==""||len(c.MinioSecretKey)<8{return c,fmt.Errorf("invalid MinIO endpoint or credentials")}
 for _,raw:=range []string{c.MediaPublicURL,c.MediaUploadPublicURL}{if raw!=""{parsed,e:=url.Parse(raw);if e!=nil||parsed.Host==""||(parsed.Scheme!="http"&&parsed.Scheme!="https"){return c,fmt.Errorf("invalid media public URL")}}}
 if c.Environment=="production" {
  local:=u.Hostname()=="localhost"||u.Hostname()=="127.0.0.1"||u.Hostname()=="::1"
  if !local&&u.Scheme=="http"&&os.Getenv("ALLOW_INSECURE_HTTP")!="true"{return c,fmt.Errorf("HTTP deployment requires ALLOW_INSECURE_HTTP=true")}
  if (u.Scheme=="https")!=c.CookieSecure{return c,fmt.Errorf("COOKIE_SECURE must match APP_ORIGIN scheme")}
 }
 if c.MediaUploadPublicURL==""{c.MediaUploadPublicURL=c.AppOrigin+"/media"};if c.MediaPublicURL==""{c.MediaPublicURL=c.AppOrigin+"/media/"+c.MinioBucket}
 return c,nil
}

func (c Config) DSN() (string,error) {
 u,err:=url.Parse(c.DatabaseURL);if err!=nil||u.Scheme!="mysql"||u.Hostname()==""||u.User==nil||strings.Trim(u.Path,"/")==""{return "",fmt.Errorf("DATABASE_URL must be a mysql:// URL")}
 password,_:=u.User.Password();port:=u.Port();if port==""{port="3306"}
 cfg:=mysql.NewConfig();cfg.User=u.User.Username();cfg.Passwd=password;cfg.Net="tcp";cfg.Addr=u.Hostname()+":"+port;cfg.DBName=strings.Trim(u.Path,"/");cfg.ParseTime=true;cfg.Loc=time.UTC;cfg.Timeout=5*time.Second;cfg.ReadTimeout=15*time.Second;cfg.WriteTimeout=15*time.Second;cfg.Params=map[string]string{"charset":"utf8mb4"};return cfg.FormatDSN(),nil
}
