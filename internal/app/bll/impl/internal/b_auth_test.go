package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"strings"
	"testing"
	"time"

	"go-gin/internal/app/config"
	httperrors "go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"
	"go-gin/pkg/auth"
)

// Unused embedded methods deliberately remain unset: these tests cannot reach
// a real database, mail service, token store or other external dependency.
type authModelStub struct {
	model.IAuth
	findActiveUser func(context.Context, string) (*schema.User, error)
	createSession  func(context.Context, *schema.Session) error
	findSession    func(context.Context, string, bool) (*schema.Session, error)
	revokeSession  func(context.Context, string, time.Time) error
}

func (s *authModelStub) FindActiveUser(ctx context.Context, id string) (*schema.User, error) {
	return s.findActiveUser(ctx, id)
}

func (s *authModelStub) CreateSession(ctx context.Context, session *schema.Session) error {
	return s.createSession(ctx, session)
}

func (s *authModelStub) FindSession(ctx context.Context, tokenHash string, lock bool) (*schema.Session, error) {
	return s.findSession(ctx, tokenHash, lock)
}

func (s *authModelStub) RevokeSession(ctx context.Context, id string, now time.Time) error {
	return s.revokeSession(ctx, id, now)
}

type templateAutherStub struct {
	auth.Auther
	parseUserID   func(string) (string, error)
	generateToken func(string) (auth.TokenInfo, error)
}

func (s *templateAutherStub) ParseUserID(token string) (string, error) {
	return s.parseUserID(token)
}

func (s *templateAutherStub) GenerateToken(id string) (auth.TokenInfo, error) {
	return s.generateToken(id)
}

type templateTokenStub struct {
	auth.TokenInfo
	accessToken string
}

func (s *templateTokenStub) GetAccessToken() string {
	return s.accessToken
}

type authTransactionStub struct {
	begins    int
	commits   int
	rollbacks int
}

func (s *authTransactionStub) Begin(context.Context) (interface{}, error) {
	s.begins++
	return s, nil
}

func (s *authTransactionStub) Commit(context.Context, interface{}) error {
	s.commits++
	return nil
}

func (s *authTransactionStub) Rollback(context.Context, interface{}) error {
	s.rollbacks++
	return nil
}

func requireAuthHTTPStatus(t *testing.T, err error, want int) {
	t.Helper()
	var httpError *httperrors.HTTPError
	if !stderrors.As(err, &httpError) {
		t.Fatalf("expected HTTP error %d, got %v", want, err)
	}
	if httpError.Status != want {
		t.Fatalf("HTTP status = %d, want %d", httpError.Status, want)
	}
}

func TestAuthAuthenticateDelegatesToTemplateAuther(t *testing.T) {
	ctx := context.Background()
	want := &schema.User{
		ID:       "active-user",
		Username: "reader",
		Role:     "ADMIN",
		Status:   "ACTIVE",
	}
	parseCalls := 0
	lookupCalls := 0
	authenticator := &templateAutherStub{
		parseUserID: func(token string) (string, error) {
			parseCalls++
			if token != "opaque-template-token" {
				t.Fatalf("ParseUserID received %q", token)
			}
			return want.ID, nil
		},
	}
	store := &authModelStub{
		findActiveUser: func(gotContext context.Context, id string) (*schema.User, error) {
			lookupCalls++
			if gotContext != ctx || id != want.ID {
				t.Fatalf("FindActiveUser did not receive the request context and parsed ID")
			}
			return want, nil
		},
	}
	service := NewAuth(store, nil, &config.Config{}, nil, authenticator)
	user, err := service.Authenticate(ctx, "opaque-template-token")
	if err != nil {
		t.Fatal(err)
	}
	if user != want || parseCalls != 1 || lookupCalls != 1 {
		t.Fatalf("authentication did not delegate once and return the current database user")
	}
}

func TestAuthAuthenticateRejectsInvalidTokensBeforeUserLookup(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		parsedID   string
		parseError error
		parseCalls int
	}{
		{
			name: "empty token",
		},
		{
			name:       "template rejects token",
			token:      "bad-token",
			parseError: auth.ErrInvalidToken,
			parseCalls: 1,
		},
		{
			name:       "template returns empty user ID",
			token:      "token-without-user",
			parseCalls: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parseCalls := 0
			authenticator := &templateAutherStub{
				parseUserID: func(string) (string, error) {
					parseCalls++
					return test.parsedID, test.parseError
				},
			}
			store := &authModelStub{
				findActiveUser: func(context.Context, string) (*schema.User, error) {
					t.Fatal("invalid token must not cause a user lookup")
					return nil, nil
				},
			}
			service := NewAuth(store, nil, &config.Config{}, nil, authenticator)
			user, err := service.Authenticate(context.Background(), test.token)
			requireAuthHTTPStatus(t, err, 401)
			if user != nil || parseCalls != test.parseCalls {
				t.Fatalf("user = %v, parser calls = %d, want nil and %d", user, parseCalls, test.parseCalls)
			}
		})
	}
}

func TestAuthAuthenticateRejectsMissingActiveUser(t *testing.T) {
	lookupCalls := 0
	store := &authModelStub{
		findActiveUser: func(context.Context, string) (*schema.User, error) {
			lookupCalls++
			// FindActiveUser's model contract excludes pending, suspended and
			// soft-deleted users; none may authenticate merely with a valid token.
			return nil, nil
		},
	}
	authenticator := &templateAutherStub{
		parseUserID: func(string) (string, error) {
			return "unavailable-user", nil
		},
	}
	service := NewAuth(store, nil, &config.Config{}, nil, authenticator)
	user, err := service.Authenticate(context.Background(), "valid-template-token")
	requireAuthHTTPStatus(t, err, 401)
	if user != nil || lookupCalls != 1 {
		t.Fatal("a valid token must still require an active database user")
	}
}

func TestAuthAuthenticatePreservesLookupError(t *testing.T) {
	wantError := stderrors.New("user lookup failed")
	store := &authModelStub{
		findActiveUser: func(context.Context, string) (*schema.User, error) {
			return nil, wantError
		},
	}
	authenticator := &templateAutherStub{
		parseUserID: func(string) (string, error) {
			return "active-user", nil
		},
	}
	service := NewAuth(store, nil, &config.Config{}, nil, authenticator)
	user, err := service.Authenticate(context.Background(), "valid-template-token")
	if user != nil || !stderrors.Is(err, wantError) {
		t.Fatalf("lookup error was not preserved: %v", err)
	}
}

func TestAuthCreateSessionDelegatesSigningAndPersistsRefresh(t *testing.T) {
	user := &schema.User{
		ID:           "session-user",
		Email:        "reader@example.test",
		Username:     "reader",
		Role:         "ADMIN",
		Status:       "ACTIVE",
		PasswordHash: "must-not-be-exposed",
	}
	meta := schema.AuthSessionMeta{
		IPAddress: "127.0.0.1",
		UserAgent: "isolated-auth-test",
	}
	cfg := &config.Config{JWTAuth: config.JWTAuth{RefreshTTLDays: 30}}
	var saved *schema.Session
	signCalls := 0
	store := &authModelStub{
		createSession: func(_ context.Context, session *schema.Session) error {
			copy := *session
			saved = &copy
			return nil
		},
	}
	authenticator := &templateAutherStub{
		generateToken: func(id string) (auth.TokenInfo, error) {
			signCalls++
			if id != user.ID {
				t.Fatalf("GenerateToken user ID = %q, want %q", id, user.ID)
			}
			return &templateTokenStub{accessToken: "template-generated-token"}, nil
		},
	}
	service := NewAuth(store, nil, cfg, nil, authenticator)
	before := time.Now().UTC()
	result, err := service.createSession(context.Background(), user, meta)
	after := time.Now().UTC()
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || saved == nil {
		t.Fatal("session result and persisted refresh session are required")
	}
	if signCalls != 1 || result.AccessToken != "template-generated-token" {
		t.Fatal("access token must come from the injected template Auther")
	}
	if result.RefreshToken == "" || result.RefreshToken == result.AccessToken {
		t.Fatal("refresh token must remain independently generated")
	}
	digest := sha256.Sum256([]byte(result.RefreshToken))
	if saved.TokenHash != hex.EncodeToString(digest[:]) || saved.TokenHash == result.RefreshToken {
		t.Fatal("only the refresh token hash may be persisted")
	}
	if saved.ID == "" || saved.UserID != user.ID || saved.RevokedAt != nil {
		t.Fatal("refresh session identity or initial revocation state changed")
	}
	if saved.CreatedAt.Before(before) || saved.CreatedAt.After(after) {
		t.Fatal("refresh session creation time is outside the request interval")
	}
	if !saved.ExpiresAt.Equal(saved.CreatedAt.Add(30 * 24 * time.Hour)) {
		t.Fatal("refresh lifetime must still use the configured number of days")
	}
	if saved.UserAgent == nil || *saved.UserAgent != meta.UserAgent ||
		saved.IPAddress == nil || *saved.IPAddress != meta.IPAddress {
		t.Fatal("refresh session request metadata was not persisted")
	}
	if result.User != schema.PublicAuthUser(user) {
		t.Fatal("public session user changed")
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{result.AccessToken, result.RefreshToken, user.PasswordHash} {
		if strings.Contains(string(encoded), secret) {
			t.Fatal("session JSON exposes a token or password hash")
		}
	}
}

func TestAuthCreateSessionSigningFailureDoesNotPersist(t *testing.T) {
	wantError := stderrors.New("template signing failed")
	store := &authModelStub{
		createSession: func(context.Context, *schema.Session) error {
			t.Fatal("signing failure must not persist a refresh session")
			return nil
		},
	}
	authenticator := &templateAutherStub{
		generateToken: func(string) (auth.TokenInfo, error) {
			return nil, wantError
		},
	}
	service := NewAuth(store, nil, &config.Config{}, nil, authenticator)
	result, err := service.createSession(context.Background(), &schema.User{ID: "user"}, schema.AuthSessionMeta{})
	if result != nil || !stderrors.Is(err, wantError) {
		t.Fatalf("signing failure was not returned without a session: %v", err)
	}
}

func TestAuthCreateSessionPersistenceFailureDoesNotReturnTokens(t *testing.T) {
	wantError := stderrors.New("refresh session write failed")
	store := &authModelStub{
		createSession: func(context.Context, *schema.Session) error {
			return wantError
		},
	}
	authenticator := &templateAutherStub{
		generateToken: func(string) (auth.TokenInfo, error) {
			return &templateTokenStub{accessToken: "signed-token"}, nil
		},
	}
	service := NewAuth(store, nil, &config.Config{}, nil, authenticator)
	result, err := service.createSession(context.Background(), &schema.User{ID: "user"}, schema.AuthSessionMeta{})
	if result != nil || !stderrors.Is(err, wantError) {
		t.Fatalf("persistence failure must not return usable session tokens: %v", err)
	}
}

func TestAuthRefreshKeepsRotationWithTemplateSigning(t *testing.T) {
	previousToken := "previous-opaque-refresh-token"
	digest := sha256.Sum256([]byte(previousToken))
	previous := &schema.Session{
		ID:        "previous-session",
		UserID:    "active-user",
		TokenHash: hex.EncodeToString(digest[:]),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	events := []string{}
	store := &authModelStub{
		findSession: func(_ context.Context, tokenHash string, lock bool) (*schema.Session, error) {
			events = append(events, "find-session")
			if tokenHash != previous.TokenHash || !lock {
				t.Fatal("refresh must look up the hashed token with a transaction lock")
			}
			return previous, nil
		},
		findActiveUser: func(_ context.Context, id string) (*schema.User, error) {
			events = append(events, "find-user")
			if id != previous.UserID {
				t.Fatal("refresh looked up the wrong session user")
			}
			return &schema.User{ID: id, Status: "ACTIVE"}, nil
		},
		revokeSession: func(_ context.Context, id string, _ time.Time) error {
			events = append(events, "revoke")
			if id != previous.ID {
				t.Fatal("refresh revoked the wrong session")
			}
			return nil
		},
		createSession: func(_ context.Context, session *schema.Session) error {
			events = append(events, "create")
			if session.UserID != previous.UserID || session.TokenHash == previous.TokenHash {
				t.Fatal("refresh must persist a new token for the same user")
			}
			return nil
		},
	}
	authenticator := &templateAutherStub{
		generateToken: func(id string) (auth.TokenInfo, error) {
			events = append(events, "sign")
			if id != previous.UserID {
				t.Fatal("replacement access token uses the wrong user")
			}
			return &templateTokenStub{accessToken: "replacement-template-token"}, nil
		},
	}
	transactions := &authTransactionStub{}
	cfg := &config.Config{JWTAuth: config.JWTAuth{RefreshTTLDays: 30}}
	service := NewAuth(store, nil, cfg, transactions, authenticator)
	result, err := service.Refresh(context.Background(), previousToken, schema.AuthSessionMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.AccessToken != "replacement-template-token" || result.RefreshToken == previousToken {
		t.Fatal("refresh did not return replacement template and refresh tokens")
	}
	if got := strings.Join(events, ","); got != "find-session,find-user,revoke,sign,create" {
		t.Fatalf("refresh sequence = %s", got)
	}
	if transactions.begins != 1 || transactions.commits != 1 || transactions.rollbacks != 0 {
		t.Fatalf("refresh transaction counts: %+v", transactions)
	}
}
