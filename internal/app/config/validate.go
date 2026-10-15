package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// setDefaults fills optional TOML settings without reading process variables.
func (c *Config) setDefaults() {
	if c.RunMode == "" {
		c.RunMode = "debug"
	}
	if c.HTTP.Host == "" {
		c.HTTP.Host = "0.0.0.0"
	}
	if c.HTTP.Port == 0 {
		c.HTTP.Port = 4000
	}
	if c.HTTP.ShutdownTimeout == 0 {
		c.HTTP.ShutdownTimeout = 15
	}
	if len(c.HTTP.TrustedProxies) == 0 {
		c.HTTP.TrustedProxies = []string{"127.0.0.1/32", "::1/128"}
	}
	if c.Store == "" {
		c.Store = "gorm"
	}
	if c.Gorm.DBType == "" {
		c.Gorm.DBType = "mysql"
	}
	if c.Gorm.MaxOpenConns == 0 {
		c.Gorm.MaxOpenConns = 8
	}
	if c.Gorm.MaxIdleConns == 0 {
		c.Gorm.MaxIdleConns = 2
	}
	if c.Gorm.MaxLifetime == 0 {
		c.Gorm.MaxLifetime = 180
	}
	if c.MySQL.Port == 0 {
		c.MySQL.Port = 3306
	}
	if c.MySQL.Parameters == "" {
		c.MySQL.Parameters = "charset=utf8mb4&parseTime=True&loc=UTC&timeout=5s&readTimeout=15s&writeTimeout=15s"
	}
	if c.JWTAuth.SigningMethod == "" {
		c.JWTAuth.SigningMethod = "HS512"
	}
	if c.JWTAuth.Expired == 0 {
		c.JWTAuth.Expired = 7200
	}
	if c.JWTAuth.RefreshTTLDays == 0 {
		c.JWTAuth.RefreshTTLDays = 30
	}
	if c.Site.Origin == "" {
		c.Site.Origin = "http://localhost:3000"
	}
	c.Site.Origin = strings.TrimRight(c.Site.Origin, "/")
	if c.Mail.Mode == "" {
		c.Mail.Mode = "log"
	}
	if c.Mail.Host == "" {
		c.Mail.Host = "localhost"
	}
	if c.Mail.Port == 0 {
		c.Mail.Port = 1025
	}
	if c.Mail.From == "" {
		c.Mail.From = "日落的足迹 <noreply@example.com>"
	}
	if c.Minio.Addr == "" {
		c.Minio.Addr = "localhost:9000"
	}
	if c.Minio.Bucket == "" {
		c.Minio.Bucket = "sunset-media"
	}
	if c.Minio.RequestTimeoutMS == 0 {
		c.Minio.RequestTimeoutMS = 5000
	}
	c.Media.PublicURL = strings.TrimRight(c.Media.PublicURL, "/")
	c.Media.UploadPublicURL = strings.TrimRight(c.Media.UploadPublicURL, "/")
	if c.Media.PublicURL == "" {
		c.Media.PublicURL = c.Site.Origin + "/media/" + c.Minio.Bucket
	}
	if c.Media.UploadPublicURL == "" {
		c.Media.UploadPublicURL = c.Site.Origin + "/media"
	}
	if c.Log.Output == "" {
		c.Log.Output = "stdout"
		c.Log.Format = "text"
		c.Log.Level = 4
	}
}

func (c *Config) ValidateDatabase() error {
	if c.Store != "gorm" || c.Gorm.DBType != "mysql" {
		return fmt.Errorf("existing site data requires store=gorm and gorm.db_type=mysql")
	}
	if c.MySQL.Host == "" || c.MySQL.User == "" || c.MySQL.DBName == "" {
		return fmt.Errorf("mysql.host, mysql.user and mysql.db_name are required")
	}
	if c.MySQL.Port < 1 || c.MySQL.Port > 65535 {
		return fmt.Errorf("mysql.port must be between 1 and 65535")
	}
	return nil
}

func (c *Config) Validate() error {
	if err := c.ValidateDatabase(); err != nil {
		return err
	}
	if c.RunMode != "debug" && c.RunMode != "release" && c.RunMode != "test" {
		return fmt.Errorf("run_mode must be debug, release or test")
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 || c.HTTP.ShutdownTimeout < 1 {
		return fmt.Errorf("invalid http.port or http.shutdown_timeout")
	}
	for _, proxy := range c.HTTP.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return fmt.Errorf("invalid http.trusted_proxies entry")
			}
		}
	}
	if c.JWTAuth.SigningKey == "" {
		return fmt.Errorf("jwt_auth.signing_key is required")
	}
	switch c.JWTAuth.SigningMethod {
	case "HS256", "HS384", "HS512":
	default:
		return fmt.Errorf("jwt_auth.signing_method must be HS256, HS384 or HS512")
	}
	if c.JWTAuth.Expired < 1 ||
		c.JWTAuth.RefreshTTLDays < 1 || c.JWTAuth.RefreshTTLDays > 3650 {
		return fmt.Errorf("invalid jwt_auth.expired or refresh_ttl_days")
	}
	if c.Mail.Mode != "log" && c.Mail.Mode != "smtp" {
		return fmt.Errorf("mail.mode must be log or smtp")
	}
	if c.Mail.Port < 1 || c.Mail.Port > 65535 {
		return fmt.Errorf("invalid mail.port")
	}
	host, port, err := net.SplitHostPort(c.Minio.Addr)
	if err != nil || host == "" {
		return fmt.Errorf("minio.addr must contain a host and port")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("invalid minio.addr port")
	}
	if c.Minio.AssessKey == "" || len(c.Minio.SecretKey) < 8 {
		return fmt.Errorf("invalid minio.assess_key or minio.secret_key")
	}
	if c.Minio.RequestTimeoutMS < 1 || c.Minio.RequestTimeoutMS > 30000 {
		return fmt.Errorf("invalid minio.request_timeout_ms")
	}
	origin, err := url.Parse(c.Site.Origin)
	if err != nil || origin.Host == "" ||
		(origin.Scheme != "http" && origin.Scheme != "https") ||
		origin.Path != "" || origin.RawQuery != "" ||
		origin.Fragment != "" || origin.User != nil {
		return fmt.Errorf("site.origin must be an HTTP(S) origin")
	}
	if c.RunMode == "release" {
		local := origin.Hostname() == "localhost" || origin.Hostname() == "127.0.0.1" || origin.Hostname() == "::1"
		if !local && origin.Scheme == "http" && !c.Site.AllowInsecureHTTP {
			return fmt.Errorf("HTTP deployment requires site.allow_insecure_http=true")
		}
		if (origin.Scheme == "https") != c.JWTAuth.CookieSecure {
			return fmt.Errorf("jwt_auth.cookie_secure must match site.origin scheme")
		}
	}
	for _, value := range []string{c.Media.PublicURL, c.Media.UploadPublicURL} {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fmt.Errorf("invalid media.public_url or media.upload_public_url")
		}
	}
	return nil
}
