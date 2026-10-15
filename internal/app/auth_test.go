package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gin/internal/app/config"
	"go-gin/pkg/auth"

	jwt "github.com/dgrijalva/jwt-go"
)

func TestInitAuthUsesTemplateJWT(t *testing.T) {
	for _, method := range []string{"HS256", "HS384", "HS512"} {
		t.Run(method, func(t *testing.T) {
			directory := t.TempDir()
			configPath := filepath.Join(directory, "config.toml")
			contents := fmt.Sprintf(`
[mysql]
host = "localhost"
user = "test"
db_name = "test"
[jwt_auth]
signing_method = %q
signing_key = "GINADMIN"
expired = 7200
store = "file"
file_path = %q
[minio]
assess_key = "test-access-key"
secret_key = "test-secret-key"
`, method, filepath.Join(directory, "jwt_auth.db"))
			if err := os.WriteFile(configPath, []byte(contents), 0600); err != nil {
				t.Fatal(err)
			}
			if err := config.LoadGlobalConfig(configPath); err != nil {
				t.Fatal(err)
			}
			auther, err := InitAuth()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := auther.Release(); err != nil {
					t.Errorf("release template JWT store: %v", err)
				}
			})

			tokenInfo, err := auther.GenerateToken("existing-user-id")
			if err != nil {
				t.Fatal(err)
			}
			if tokenInfo.GetTokenType() != "Bearer" {
				t.Fatalf("unexpected token type %q", tokenInfo.GetTokenType())
			}
			remaining := time.Until(time.Unix(tokenInfo.GetExpiresAt(), 0))
			if remaining < 7190*time.Second || remaining > 7200*time.Second {
				t.Fatalf("JWT lifetime does not match config: %v", remaining)
			}
			rawToken := tokenInfo.GetAccessToken()
			claims := &jwt.StandardClaims{}
			parsed, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
				if token.Method.Alg() != method {
					return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
				}
				return []byte("GINADMIN"), nil
			})
			if err != nil || !parsed.Valid || claims.Subject != "existing-user-id" {
				t.Fatalf("token does not use template claims/signature: %v", err)
			}
			userID, err := auther.ParseUserID(rawToken)
			if err != nil || userID != "existing-user-id" {
				t.Fatalf("template token could not be authenticated: user=%q err=%v", userID, err)
			}

			invalidClaims := jwt.StandardClaims{
				Subject:   "existing-user-id",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			}
			wrongSignature, err := jwt.NewWithClaims(jwt.GetSigningMethod(method), invalidClaims).SignedString([]byte("wrong-key"))
			if err != nil {
				t.Fatal(err)
			}
			invalidClaims.ExpiresAt = time.Now().Add(-time.Hour).Unix()
			expiredToken, err := jwt.NewWithClaims(jwt.GetSigningMethod(method), invalidClaims).SignedString([]byte("GINADMIN"))
			if err != nil {
				t.Fatal(err)
			}
			for _, invalid := range []string{"not-a-jwt", wrongSignature, expiredToken} {
				if _, err := auther.ParseUserID(invalid); !errors.Is(err, auth.ErrInvalidToken) {
					t.Fatalf("invalid token was not rejected: %v", err)
				}
			}

			if err := auther.DestroyToken(rawToken); err != nil {
				t.Fatal(err)
			}
			if _, err := auther.ParseUserID(rawToken); !errors.Is(err, auth.ErrInvalidToken) {
				t.Fatalf("template file store did not reject revoked token: %v", err)
			}
		})
	}
}
