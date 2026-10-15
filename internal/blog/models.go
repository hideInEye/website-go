package blog

import "time"

// Explicit camelCase columns preserve the existing Prisma database contract.
type User struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 Email string `gorm:"column:email" json:"email"`
 Username string `gorm:"column:username" json:"username"`
 DisplayName string `gorm:"column:displayName" json:"displayName"`
 AvatarURL *string `gorm:"column:avatarUrl" json:"avatarUrl"`
 Bio *string `gorm:"column:bio" json:"bio"`
 PasswordHash string `gorm:"column:passwordHash" json:"-"`
 Role string `gorm:"column:role" json:"role"`
 Status string `gorm:"column:status" json:"status"`
 EmailVerifiedAt *time.Time `gorm:"column:emailVerifiedAt" json:"emailVerifiedAt"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
 UpdatedAt time.Time `gorm:"column:updatedAt" json:"updatedAt"`
 DeletedAt *time.Time `gorm:"column:deletedAt" json:"-"`
}
func (User) TableName() string { return "users" }

type Session struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 TokenHash string `gorm:"column:tokenHash" json:"-"`
 UserID string `gorm:"column:userId" json:"userId"`
 UserAgent *string `gorm:"column:userAgent" json:"userAgent"`
 IPAddress *string `gorm:"column:ipAddress" json:"ipAddress"`
 ExpiresAt time.Time `gorm:"column:expiresAt" json:"expiresAt"`
 RevokedAt *time.Time `gorm:"column:revokedAt" json:"revokedAt"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
}
func (Session) TableName() string { return "sessions" }

type VerificationToken struct {
 ID string `gorm:"primary_key;column:id"`
 TokenHash string `gorm:"column:tokenHash"`
 UserID string `gorm:"column:userId"`
 ExpiresAt time.Time `gorm:"column:expiresAt"`
 UsedAt *time.Time `gorm:"column:usedAt"`
 CreatedAt time.Time `gorm:"column:createdAt"`
}
func (VerificationToken) TableName() string { return "email_verification_tokens" }
type ResetToken VerificationToken
func (ResetToken) TableName() string { return "password_reset_tokens" }

type Category struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 Name string `gorm:"column:name" json:"name"`
 Slug string `gorm:"column:slug" json:"slug"`
 Description *string `gorm:"column:description" json:"description"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
 UpdatedAt time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}
func (Category) TableName() string { return "categories" }
type Tag struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 Name string `gorm:"column:name" json:"name"`
 Slug string `gorm:"column:slug" json:"slug"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
 UpdatedAt time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}
func (Tag) TableName() string { return "tags" }
type Post struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 LegacyRecordID *string `gorm:"column:legacyRecordId" json:"legacyRecordId"`
 Slug string `gorm:"column:slug" json:"slug"`
 Title string `gorm:"column:title" json:"title"`
 Excerpt string `gorm:"column:excerpt" json:"excerpt"`
 Content string `gorm:"column:content" json:"content"`
 CoverURL *string `gorm:"column:coverUrl" json:"coverUrl"`
 Status string `gorm:"column:status" json:"status"`
 SEOTitle *string `gorm:"column:seoTitle" json:"seoTitle"`
 SEODescription *string `gorm:"column:seoDescription" json:"seoDescription"`
 AuthorID string `gorm:"column:authorId" json:"authorId"`
 CategoryID string `gorm:"column:categoryId" json:"categoryId"`
 PublishedAt *time.Time `gorm:"column:publishedAt" json:"publishedAt"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
 UpdatedAt time.Time `gorm:"column:updatedAt" json:"updatedAt"`
 DeletedAt *time.Time `gorm:"column:deletedAt" json:"-"`
}
func (Post) TableName() string { return "posts" }
type PostTag struct {
 PostID string `gorm:"primary_key;column:postId" json:"postId"`
 TagID string `gorm:"primary_key;column:tagId" json:"tagId"`
}
func (PostTag) TableName() string { return "post_tags" }
type SlugHistory struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 PostID string `gorm:"column:postId" json:"postId"`
 Slug string `gorm:"column:slug" json:"slug"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
}
func (SlugHistory) TableName() string { return "post_slug_history" }
type Comment struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 Content string `gorm:"column:content" json:"content"`
 Status string `gorm:"column:status" json:"status"`
 PostID string `gorm:"column:postId" json:"postId"`
 UserID string `gorm:"column:userId" json:"userId"`
 ParentID *string `gorm:"column:parentId" json:"parentId"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
 UpdatedAt time.Time `gorm:"column:updatedAt" json:"updatedAt"`
 DeletedAt *time.Time `gorm:"column:deletedAt" json:"-"`
}
func (Comment) TableName() string { return "comments" }
type Bookmark struct {
 UserID string `gorm:"primary_key;column:userId" json:"userId"`
 PostID string `gorm:"primary_key;column:postId" json:"postId"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
}
func (Bookmark) TableName() string { return "bookmarks" }
type Media struct {
 ID string `gorm:"primary_key;column:id" json:"id"`
 ObjectKey string `gorm:"column:objectKey" json:"objectKey"`
 OriginalName string `gorm:"column:originalName" json:"originalName"`
 MimeType string `gorm:"column:mimeType" json:"mimeType"`
 Size int64 `gorm:"column:size" json:"size"`
 Width *int `gorm:"column:width" json:"width"`
 Height *int `gorm:"column:height" json:"height"`
 AltText *string `gorm:"column:altText" json:"altText"`
 Status string `gorm:"column:status" json:"status"`
 UploaderID string `gorm:"column:uploaderId" json:"uploaderId"`
 CreatedAt time.Time `gorm:"column:createdAt" json:"createdAt"`
 UpdatedAt time.Time `gorm:"column:updatedAt" json:"updatedAt"`
}
func (Media) TableName() string { return "media" }
