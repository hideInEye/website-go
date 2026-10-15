package model

import (
	"context"
	"time"

	"go-gin/internal/app/schema"
)

// IAuth owns persistence only. Transactions are resolved from context through
// the template's existing ITrans and ExecTrans mechanism.
type IAuth interface {
	FindUserByIdentity(context.Context, string, string) (*schema.User, error)
	FindUserByEmail(context.Context, string) (*schema.User, error)
	FindActiveUser(context.Context, string) (*schema.User, error)
	CreateUser(context.Context, *schema.User) error
	ActivateUser(context.Context, string, time.Time) (bool, error)
	UpdatePassword(context.Context, string, string, time.Time) (bool, error)
	CreateVerification(context.Context, *schema.VerificationToken) error
	FindVerification(context.Context, string, bool) (*schema.VerificationToken, error)
	ConsumeVerification(context.Context, string, time.Time) (bool, error)
	CreateResetToken(context.Context, *schema.ResetToken) error
	FindResetToken(context.Context, string) (*schema.ResetToken, error)
	ConsumeResetToken(context.Context, string, time.Time) (bool, error)
	CreateSession(context.Context, *schema.Session) error
	FindSession(context.Context, string, bool) (*schema.Session, error)
	RevokeSession(context.Context, string, time.Time) error
	RevokeSessionToken(context.Context, string, time.Time) error
	RevokeUserSessions(context.Context, string, time.Time) error
}
