package internal

import (
	"context"
	"strings"
	"time"

	contract "go-gin/internal/app/bll"
	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"

	"github.com/google/uuid"
)

type Comment struct {
	CommentModel model.IComment
}

var _ contract.IComment = (*Comment)(nil)

func NewComment(comments model.IComment) *Comment {
	return &Comment{CommentModel: comments}
}

func commentObject(comment schema.Comment) schema.Object {
	return schema.Object{
		"id":        comment.ID,
		"content":   comment.Content,
		"status":    comment.Status,
		"postId":    comment.PostID,
		"userId":    comment.UserID,
		"parentId":  comment.ParentID,
		"createdAt": comment.CreatedAt,
		"updatedAt": comment.UpdatedAt,
		"deletedAt": comment.DeletedAt,
	}
}

func commentUser(user schema.User) schema.Object {
	return schema.Object{
		"username":    user.Username,
		"displayName": user.DisplayName,
		"avatarUrl":   user.AvatarURL,
	}
}

func (a *Comment) PublicComments(ctx context.Context, postID string) ([]schema.Object, error) {
	roots, err := a.CommentModel.PublicRoots(ctx, postID)
	if err != nil {
		return nil, err
	}
	items := make([]schema.Object, 0, len(roots))
	if len(roots) == 0 {
		return items, nil
	}

	parentIDs := make([]string, 0, len(roots))
	for _, root := range roots {
		parentIDs = append(parentIDs, root.ID)
	}
	replies, err := a.CommentModel.PublicReplies(ctx, parentIDs)
	if err != nil {
		return nil, err
	}
	all := make([]schema.Comment, 0, len(roots)+len(replies))
	all = append(all, roots...)
	all = append(all, replies...)
	users, _, err := a.CommentModel.Relations(ctx, all)
	if err != nil {
		return nil, err
	}

	repliesByParent := make(map[string][]schema.Object)
	for _, reply := range replies {
		item := commentObject(reply)
		item["user"] = commentUser(users[reply.UserID])
		if reply.ParentID != nil {
			repliesByParent[*reply.ParentID] = append(repliesByParent[*reply.ParentID], item)
		}
	}
	for _, root := range roots {
		item := commentObject(root)
		item["user"] = commentUser(users[root.UserID])
		children := repliesByParent[root.ID]
		if children == nil {
			children = []schema.Object{}
		}
		item["replies"] = children
		items = append(items, item)
	}
	return items, nil
}

func (a *Comment) MyComments(ctx context.Context, user schema.User) ([]schema.Object, error) {
	comments, err := a.CommentModel.ByUser(ctx, user.ID, "", 0)
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
			"status":    comment.Status,
			"createdAt": comment.CreatedAt,
			"post":      schema.Object{"title": post.Title, "slug": post.Slug},
		})
	}
	return items, nil
}

func (a *Comment) CreateComment(ctx context.Context, user schema.User, input schema.CommentInput) (schema.Object, error) {
	if !input.PostID.Valid(0, 0, true, false) ||
		!input.ParentID.Valid(0, 0, false, true) ||
		!input.Content.Valid(1, 1000, true, false) {
		return nil, errors.NewHTTP(400, "评论参数不正确")
	}

	postID := input.PostID.Text("")
	exists, err := a.CommentModel.PublishedPostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewHTTP(404, "文章不存在")
	}
	if parentID := input.ParentID.Text(""); parentID != "" {
		parent, err := a.CommentModel.FindReplyParent(ctx, parentID, postID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.NewHTTP(400, "只能回复已发布的一级评论")
		}
	}

	approved, err := a.CommentModel.CountApproved(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	status := "PENDING"
	if user.Role == "ADMIN" || approved > 0 {
		status = "APPROVED"
	}
	now := time.Now().UTC()
	comment := schema.Comment{
		ID:        strings.ReplaceAll(uuid.NewString(), "-", ""),
		Content:   strings.TrimSpace(input.Content.Text("")),
		Status:    status,
		PostID:    postID,
		UserID:    user.ID,
		ParentID:  input.ParentID.Value,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := a.CommentModel.Create(ctx, comment); err != nil {
		return nil, err
	}
	item := commentObject(comment)
	item["moderationRequired"] = status == "PENDING"
	return item, nil
}

func (a *Comment) DeleteComment(ctx context.Context, user schema.User, id string) (schema.Object, error) {
	comment, err := a.CommentModel.FindIncludingDeleted(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, errors.NewHTTP(404, "评论不存在")
	}
	if comment.UserID != user.ID && user.Role != "ADMIN" {
		return nil, errors.NewHTTP(403, "只能删除自己的评论")
	}
	if err := a.CommentModel.SoftDelete(ctx, id, time.Now().UTC()); err != nil {
		return nil, err
	}
	return schema.Object{"deleted": true}, nil
}

func (a *Comment) PendingComments(ctx context.Context, user schema.User) ([]schema.Object, error) {
	if user.Role != "ADMIN" {
		return nil, errors.NewHTTP(403, "需要管理员权限")
	}
	comments, err := a.CommentModel.Pending(ctx)
	if err != nil {
		return nil, err
	}
	users, posts, err := a.CommentModel.Relations(ctx, comments)
	if err != nil {
		return nil, err
	}
	items := make([]schema.Object, 0, len(comments))
	for _, comment := range comments {
		item := commentObject(comment)
		author, post := users[comment.UserID], posts[comment.PostID]
		item["user"] = schema.Object{"username": author.Username, "displayName": author.DisplayName}
		item["post"] = schema.Object{"title": post.Title, "slug": post.Slug}
		items = append(items, item)
	}
	return items, nil
}

func (a *Comment) ModerateComment(ctx context.Context, user schema.User, id string, input schema.CommentStatusInput) (schema.Object, error) {
	if user.Role != "ADMIN" {
		return nil, errors.NewHTTP(403, "需要管理员权限")
	}
	if input.Status != "APPROVED" && input.Status != "REJECTED" {
		return nil, errors.NewHTTP(400, "评论状态不正确")
	}
	comment, err := a.CommentModel.SetStatus(ctx, id, input.Status)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, errors.NewHTTP(404, "评论不存在")
	}
	return commentObject(*comment), nil
}
