package schema

type AuthLoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthRegisterInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthTokenInput struct {
	Token string `json:"token"`
}

type AuthForgotPasswordInput struct {
	Email string `json:"email"`
}

type AuthResetPasswordInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type AuthSessionMeta struct {
	IPAddress string
	UserAgent string
}

type AuthUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type AuthRegistration struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// Tokens are delivered only through HttpOnly cookies by the HTTP controller.
type AuthSessionResult struct {
	User         AuthUser `json:"user"`
	AccessToken  string   `json:"-"`
	RefreshToken string   `json:"-"`
}

func PublicAuthUser(user *User) AuthUser {
	return AuthUser{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}
}
