package bll

import (
	"context"

	"go-gin/internal/app/schema"
)

type IAuth interface {
	Register(context.Context, schema.AuthRegisterInput) (*schema.AuthRegistration, error)
	VerifyEmail(context.Context, string) error
	Login(context.Context, schema.AuthLoginInput, schema.AuthSessionMeta) (*schema.AuthSessionResult, error)
	Refresh(context.Context, string, schema.AuthSessionMeta) (*schema.AuthSessionResult, error)
	Logout(context.Context, string) error
	ForgotPassword(context.Context, schema.AuthForgotPasswordInput) error
	ResetPassword(context.Context, schema.AuthResetPasswordInput) error
	Authenticate(context.Context, string) (*schema.User, error)
}

type IMail interface {
	SendVerification(context.Context, string, string) error
	SendPasswordReset(context.Context, string, string) error
}
