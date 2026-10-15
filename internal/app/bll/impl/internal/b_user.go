package internal

import (
	"context"
	"regexp"
	"time"

	contract "go-gin/internal/app/bll"
	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"
)

type User struct {
	UserModel    model.IUser
	CommentModel model.IComment
}

var _ contract.IUser = (*User)(nil)

var profileUsername = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func NewUser(users model.IUser, comments model.IComment) *User {
	return &User{UserModel: users, CommentModel: comments}
}

func profileObject(user schema.User) schema.Object {
	return schema.Object{
		"id":              user.ID,
		"email":           user.Email,
		"role":            user.Role,
		"status":          user.Status,
		"emailVerifiedAt": user.EmailVerifiedAt,
		"username":        user.Username,
		"displayName":     user.DisplayName,
		"avatarUrl":       user.AvatarURL,
		"bio":             user.Bio,
		"createdAt":       user.CreatedAt,
	}
}

func (a *User) PublicProfile(ctx context.Context, username string) (schema.Object, error) {
	user, err := a.UserModel.FindPublic(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewHTTP(404, "用户不存在")
	}

	comments, err := a.CommentModel.ByUser(ctx, user.ID, "APPROVED", 20)
	if err != nil {
		return nil, err
	}
	_, posts, err := a.CommentModel.Relations(ctx, comments)
	if err != nil {
		return nil, err
	}
	items := make([]schema.Object, 0, len(comments))
	for _, comment := range comments {
		post := posts[comment.PostID]
		items = append(items, schema.Object{
			"id":        comment.ID,
			"content":   comment.Content,
			"createdAt": comment.CreatedAt,
			"post":      schema.Object{"title": post.Title, "slug": post.Slug},
		})
	}
	return schema.Object{
		"username":    user.Username,
		"displayName": user.DisplayName,
		"avatarUrl":   user.AvatarURL,
		"bio":         user.Bio,
		"createdAt":   user.CreatedAt,
		"comments":    items,
	}, nil
}

func (a *User) MyProfile(ctx context.Context, user schema.User) (schema.Object, error) {
	return profileObject(user), nil
}

func (a *User) UpdateProfile(ctx context.Context, actor schema.User, input schema.ProfileInput) (schema.Object, error) {
	validUsername := !input.Username.Present || profileUsername.MatchString(input.Username.Text(""))
	if !input.DisplayName.Valid(2, 80, false, false) ||
		!input.Username.Valid(2, 32, false, false) || !validUsername ||
		!input.Bio.Valid(0, 500, false, true) ||
		!input.AvatarURL.Valid(0, 2048, false, true) ||
		!schema.ValidContentURL(input.AvatarURL.Value) {
		return nil, errors.NewHTTP(400, "个人资料参数不正确")
	}

	updates := map[string]interface{}{"updatedAt": time.Now().UTC()}
	fields := map[string]schema.OptionalString{
		"displayName": input.DisplayName,
		"username":    input.Username,
		"bio":         input.Bio,
		"avatarUrl":   input.AvatarURL,
	}
	for name, field := range fields {
		if field.Present {
			if field.Value == nil {
				updates[name] = nil
			} else {
				updates[name] = *field.Value
			}
		}
	}
	user, err := a.UserModel.UpdateProfile(ctx, actor.ID, updates)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewHTTP(404, "用户不存在")
	}
	return profileObject(*user), nil
}

func (a *User) ListUsers(ctx context.Context, actor schema.User) ([]schema.Object, error) {
	if actor.Role != "ADMIN" {
		return nil, errors.NewHTTP(403, "需要管理员权限")
	}
	users, err := a.UserModel.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]schema.Object, 0, len(users))
	for _, user := range users {
		items = append(items, schema.Object{
			"id":          user.ID,
			"email":       user.Email,
			"username":    user.Username,
			"displayName": user.DisplayName,
			"role":        user.Role,
			"status":      user.Status,
			"createdAt":   user.CreatedAt,
		})
	}
	return items, nil
}

func (a *User) SetUserStatus(ctx context.Context, actor schema.User, id string, input schema.UserStatusInput) (schema.Object, error) {
	if actor.Role != "ADMIN" {
		return nil, errors.NewHTTP(403, "需要管理员权限")
	}
	if input.Status != "ACTIVE" && input.Status != "SUSPENDED" {
		return nil, errors.NewHTTP(400, "用户状态不正确")
	}
	if actor.ID == id && input.Status == "SUSPENDED" {
		return nil, errors.NewHTTP(400, "不能暂停自己的管理员账号")
	}
	user, err := a.UserModel.SetStatus(ctx, id, input.Status)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewHTTP(404, "用户不存在")
	}
	return schema.Object{"id": user.ID, "username": user.Username, "status": user.Status}, nil
}
