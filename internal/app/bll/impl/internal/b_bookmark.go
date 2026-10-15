package internal

import (
	"context"

	contract "go-gin/internal/app/bll"
	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"
)

type Bookmark struct {
	BookmarkModel model.IBookmark
	PostModel     model.IPost
}

var _ contract.IBookmark = (*Bookmark)(nil)

func NewBookmark(bookmarks model.IBookmark, posts model.IPost) *Bookmark {
	return &Bookmark{BookmarkModel: bookmarks, PostModel: posts}
}

func (a *Bookmark) HasBookmark(ctx context.Context, user schema.User, postID string) (schema.Object, error) {
	exists, err := a.BookmarkModel.Exists(ctx, user.ID, postID)
	if err != nil {
		return nil, err
	}
	return schema.Object{"bookmarked": exists}, nil
}

func (a *Bookmark) ListBookmarks(ctx context.Context, user schema.User) ([]schema.Object, error) {
	posts, err := a.BookmarkModel.ListPosts(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return a.PostModel.PresentPosts(ctx, posts, true, true)
}

func (a *Bookmark) AddBookmark(ctx context.Context, user schema.User, postID string) (schema.Object, error) {
	exists, err := a.BookmarkModel.PublishedPostExists(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewHTTP(404, "文章不存在")
	}
	if err := a.BookmarkModel.Add(ctx, user.ID, postID); err != nil {
		return nil, err
	}
	return schema.Object{"bookmarked": true}, nil
}

func (a *Bookmark) RemoveBookmark(ctx context.Context, user schema.User, postID string) (schema.Object, error) {
	if err := a.BookmarkModel.Remove(ctx, user.ID, postID); err != nil {
		return nil, err
	}
	return schema.Object{"bookmarked": false}, nil
}
