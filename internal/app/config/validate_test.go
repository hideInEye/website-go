package config

import "testing"

func validJWTConfig() Config {
	var cfg Config
	cfg.MySQL.Host = "localhost"
	cfg.MySQL.User = "test"
	cfg.MySQL.DBName = "test"
	cfg.JWTAuth.SigningKey = "GINADMIN"
	cfg.Minio.AssessKey = "test-access-key"
	cfg.Minio.SecretKey = "test-secret-key"
	cfg.setDefaults()
	return cfg
}

func TestJWTDefaultsMatchTemplate(t *testing.T) {
	cfg := validJWTConfig()
	if cfg.JWTAuth.SigningMethod != "HS512" || cfg.JWTAuth.Expired != 7200 {
		t.Fatalf("unexpected JWT defaults: method=%s expired=%d", cfg.JWTAuth.SigningMethod, cfg.JWTAuth.Expired)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("template JWT configuration was rejected: %v", err)
	}
}

func TestJWTValidationSupportsTemplateMethods(t *testing.T) {
	for _, method := range []string{"HS256", "HS384", "HS512"} {
		t.Run(method, func(t *testing.T) {
			cfg := validJWTConfig()
			cfg.JWTAuth.SigningMethod = method
			if err := cfg.Validate(); err != nil {
				t.Fatalf("supported template method was rejected: %v", err)
			}
		})
	}
}

func TestJWTValidationRejectsInvalidSettings(t *testing.T) {
	tests := []struct {
		name   string
		change func(*JWTAuth)
	}{
		{"empty key", func(cfg *JWTAuth) { cfg.SigningKey = "" }},
		{"unsigned tokens", func(cfg *JWTAuth) { cfg.SigningMethod = "none" }},
		{"unsupported method", func(cfg *JWTAuth) { cfg.SigningMethod = "RS256" }},
		{"invalid expiry", func(cfg *JWTAuth) { cfg.Expired = -1 }},
		{"invalid refresh lifetime", func(cfg *JWTAuth) { cfg.RefreshTTLDays = -1 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validJWTConfig()
			test.change(&cfg.JWTAuth)
			if err := cfg.Validate(); err == nil {
				t.Fatal("invalid JWT configuration was accepted")
			}
		})
	}
}

func TestJWTDefaultsPreserveExplicitConfiguration(t *testing.T) {
	var cfg Config
	cfg.JWTAuth.SigningMethod = "HS384"
	cfg.JWTAuth.SigningKey = "custom-key"
	cfg.JWTAuth.Expired = 3600
	cfg.setDefaults()
	if cfg.JWTAuth.SigningMethod != "HS384" || cfg.JWTAuth.SigningKey != "custom-key" || cfg.JWTAuth.Expired != 3600 {
		t.Fatal("defaults replaced explicit JWT configuration")
	}
}
