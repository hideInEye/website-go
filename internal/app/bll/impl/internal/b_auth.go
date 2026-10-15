package internal

import (
	"context"
	"log"
	"strings"
	"time"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/config"
	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"
	"go-gin/pkg/auth"
)

func NewAuth(store model.IAuth, mail bll.IMail, cfg *config.Config, transactions model.ITrans, authenticator auth.Auther) *Auth {
	return &Auth{
		store:         store,
		mail:          mail,
		config:        cfg,
		transactions:  transactions,
		authenticator: authenticator,
	}
}

type Auth struct {
	store         model.IAuth
	mail          bll.IMail
	config        *config.Config
	transactions  model.ITrans
	authenticator auth.Auther
}

var _ bll.IAuth = (*Auth)(nil)

func (a *Auth) Authenticate(ctx context.Context, rawToken string) (*schema.User, error) {
	if rawToken == "" {
		return nil, errors.NewHTTP(401, "请先登录")
	}
	userID, err := a.authenticator.ParseUserID(rawToken)
	if err != nil || userID == "" {
		return nil, errors.NewHTTP(401, "登录状态已失效")
	}
	user, err := a.store.FindActiveUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewHTTP(401, "账号不可用")
	}
	return user, nil
}

func (a *Auth) Register(ctx context.Context, input schema.AuthRegisterInput) (*schema.AuthRegistration, error) {
	input.Email = normalizeEmail(input.Email)
	if !validEmail(input.Email) {
		return nil, errors.NewHTTP(400, "请输入有效邮箱")
	}
	if !usernamePattern.MatchString(input.Username) {
		return nil, errors.NewHTTP(400, "用户名须为 2–32 位，只能包含字母、数字、横线和下划线")
	}
	if !validPassword(input.Password, 10) {
		return nil, errors.NewHTTP(400, "密码须为 10–128 位")
	}
	existing, err := a.store.FindUserByIdentity(ctx, input.Email, input.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if strings.EqualFold(existing.Email, input.Email) {
			return nil, errors.NewHTTP(409, "该邮箱已经注册")
		}
		return nil, errors.NewHTTP(409, "该用户名已被使用")
	}
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	rawToken, err := randomAuthToken()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	user := schema.User{
		ID:           newAuthID(),
		Email:        input.Email,
		Username:     input.Username,
		DisplayName:  input.Username,
		PasswordHash: passwordHash,
		Role:         "USER",
		Status:       "PENDING_VERIFICATION",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err = ExecTrans(ctx, a.transactions, func(ctx context.Context) error {
		if err := a.store.CreateUser(ctx, &user); err != nil {
			return err
		}
		token := schema.VerificationToken{
			ID:        newAuthID(),
			TokenHash: hashAuthToken(rawToken),
			UserID:    user.ID,
			ExpiresAt: now.Add(24 * time.Hour),
			CreatedAt: now,
		}
		if err := a.store.CreateVerification(ctx, &token); err != nil {
			return err
		}
		// The transaction commits only after delivery acceptance. SMTP errors
		// therefore cannot leave an unusable account or verification row.
		if err := a.mail.SendVerification(ctx, user.Email, rawToken); err != nil {
			log.Printf("registration email delivery failed: %v", err)
			return errors.NewHTTP(500, "验证邮件发送失败，请稍后重试")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &schema.AuthRegistration{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}, nil
}

func (a *Auth) VerifyEmail(ctx context.Context, rawToken string) error {
	if !validAuthToken(rawToken) {
		return errors.NewHTTP(400, "验证令牌格式不正确")
	}
	return ExecTrans(ctx, a.transactions, func(ctx context.Context) error {
		token, err := a.store.FindVerification(ctx, hashAuthToken(rawToken), true)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if token == nil || token.UsedAt != nil || !token.ExpiresAt.After(now) {
			return errors.NewHTTP(401, "验证链接无效或已经过期")
		}
		consumed, err := a.store.ConsumeVerification(ctx, token.ID, now)
		if err != nil {
			return err
		}
		if !consumed {
			return errors.NewHTTP(401, "验证链接无效或已经过期")
		}
		updated, err := a.store.ActivateUser(ctx, token.UserID, now)
		if err != nil {
			return err
		}
		if !updated {
			return errors.NewHTTP(401, "账号不可用")
		}
		return nil
	})
}

func (a *Auth) Login(ctx context.Context, input schema.AuthLoginInput, meta schema.AuthSessionMeta) (*schema.AuthSessionResult, error) {
	input.Email = normalizeEmail(input.Email)
	if !validEmail(input.Email) || !validPassword(input.Password, 8) {
		return nil, errors.NewHTTP(400, "请输入有效邮箱及 8–128 位密码")
	}
	user, err := a.store.FindUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil || !verifyPassword(user.PasswordHash, input.Password) {
		return nil, errors.NewHTTP(401, "邮箱或密码不正确")
	}
	if user.Status == "PENDING_VERIFICATION" {
		return nil, errors.NewHTTP(401, "请先完成邮箱验证")
	}
	if user.Status != "ACTIVE" {
		return nil, errors.NewHTTP(401, "账号已被暂停")
	}
	return a.createSession(ctx, user, meta)
}

func (a *Auth) Refresh(ctx context.Context, rawToken string, meta schema.AuthSessionMeta) (*schema.AuthSessionResult, error) {
	if rawToken == "" || len(rawToken) > 512 {
		return nil, errors.NewHTTP(401, "刷新令牌无效")
	}
	var replacement *schema.AuthSessionResult
	err := ExecTrans(ctx, a.transactions, func(ctx context.Context) error {
		previous, err := a.store.FindSession(ctx, hashAuthToken(rawToken), true)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if previous == nil || previous.RevokedAt != nil || !previous.ExpiresAt.After(now) {
			return errors.NewHTTP(401, "刷新令牌无效")
		}
		user, err := a.store.FindActiveUser(ctx, previous.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return errors.NewHTTP(401, "刷新令牌无效")
		}
		if err := a.store.RevokeSession(ctx, previous.ID, now); err != nil {
			return err
		}
		replacement, err = a.createSession(ctx, user, meta)
		return err
	})
	return replacement, err
}

func (a *Auth) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return a.store.RevokeSessionToken(ctx, hashAuthToken(rawToken), time.Now().UTC())
}

func (a *Auth) ForgotPassword(ctx context.Context, input schema.AuthForgotPasswordInput) error {
	input.Email = normalizeEmail(input.Email)
	if !validEmail(input.Email) {
		return errors.NewHTTP(400, "请输入有效邮箱")
	}
	user, err := a.store.FindUserByEmail(ctx, input.Email)
	if err != nil || user == nil {
		return err
	}
	rawToken, err := randomAuthToken()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	token := schema.ResetToken{
		ID:        newAuthID(),
		TokenHash: hashAuthToken(rawToken),
		UserID:    user.ID,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}
	if err := a.store.CreateResetToken(ctx, &token); err != nil {
		return err
	}
	if err := a.mail.SendPasswordReset(ctx, user.Email, rawToken); err != nil {
		log.Printf("password reset email delivery failed: %v", err)
		return errors.NewHTTP(500, "重置邮件发送失败，请稍后重试")
	}
	return nil
}

func (a *Auth) ResetPassword(ctx context.Context, input schema.AuthResetPasswordInput) error {
	if !validAuthToken(input.Token) || !validPassword(input.Password, 10) {
		return errors.NewHTTP(400, "请输入有效令牌及 10–128 位密码")
	}
	token, err := a.store.FindResetToken(ctx, hashAuthToken(input.Token))
	if err != nil {
		return err
	}
	if token == nil || token.UsedAt != nil || !token.ExpiresAt.After(time.Now().UTC()) {
		return errors.NewHTTP(401, "重置链接无效或已经过期")
	}
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return err
	}
	return ExecTrans(ctx, a.transactions, func(ctx context.Context) error {
		now := time.Now().UTC()
		consumed, err := a.store.ConsumeResetToken(ctx, token.ID, now)
		if err != nil {
			return err
		}
		if !consumed {
			return errors.NewHTTP(401, "重置链接无效或已经过期")
		}
		updated, err := a.store.UpdatePassword(ctx, token.UserID, passwordHash, now)
		if err != nil {
			return err
		}
		if !updated {
			return errors.NewHTTP(401, "账号不可用")
		}
		return a.store.RevokeUserSessions(ctx, token.UserID, now)
	})
}

func (a *Auth) createSession(ctx context.Context, user *schema.User, meta schema.AuthSessionMeta) (*schema.AuthSessionResult, error) {
	refreshToken, err := randomAuthToken()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	token, err := a.authenticator.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}
	session := schema.Session{
		ID:        newAuthID(),
		TokenHash: hashAuthToken(refreshToken),
		UserID:    user.ID,
		UserAgent: limitedAuthString(meta.UserAgent, 512),
		IPAddress: limitedAuthString(meta.IPAddress, 64),
		ExpiresAt: now.Add(time.Duration(a.config.JWTAuth.RefreshTTLDays) * 24 * time.Hour),
		CreatedAt: now,
	}
	if err := a.store.CreateSession(ctx, &session); err != nil {
		return nil, err
	}
	return &schema.AuthSessionResult{
		User:         schema.PublicAuthUser(user),
		AccessToken:  token.GetAccessToken(),
		RefreshToken: refreshToken,
	}, nil
}
