package internal

import (
	"context"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"go-gin/internal/app/bll"
	"go-gin/internal/app/config"
	"go-gin/internal/app/errors"
	"go-gin/internal/app/model"
	"go-gin/internal/app/schema"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func NewMedia(store model.IMedia, storage *minio.Client, cfg *config.Config) *Media {
	return &Media{
		store:   store,
		storage: storage,
		config:  cfg,
	}
}

type Media struct {
	store       model.IMedia
	storage     *minio.Client
	config      *config.Config
	bucketMu    sync.Mutex
	bucketReady bool
}

var _ bll.IMedia = (*Media)(nil)

var unsafeMediaSuffix = regexp.MustCompile(`[^.a-z0-9]`)

func publicUploadURL(signed *url.URL, publicBase string) (string, error) {
	external, err := url.Parse(publicBase)
	if err != nil {
		return "", err
	}
	if external.Host == "" || (external.Scheme != "http" && external.Scheme != "https") {
		return "", errors.NewHTTP(503, "媒体上传地址配置错误")
	}
	// The proxy restores the signed MinIO Host and strips its public prefix.
	// Preserve the signed query verbatim; internal Docker names never go to UI.
	external.Path = strings.TrimRight(external.Path, "/") + signed.Path
	external.RawPath = ""
	external.RawQuery = signed.RawQuery
	external.Fragment = ""
	return external.String(), nil
}

func (a *Media) ensureBucket(ctx context.Context) error {
	a.bucketMu.Lock()
	defer a.bucketMu.Unlock()
	if a.bucketReady {
		return nil
	}
	exists, err := a.storage.BucketExists(ctx, a.config.Minio.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		err = a.storage.MakeBucket(ctx, a.config.Minio.Bucket, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil && minio.ToErrorResponse(err).Code != "BucketAlreadyOwnedByYou" {
			return err
		}
	}
	a.bucketReady = true
	return nil
}

func (a *Media) Presign(ctx context.Context, uploaderID string, input schema.MediaPresignInput) (*schema.MediaPresignResult, error) {
	if !input.Valid() {
		return nil, errors.NewHTTP(400, "图片文件名、类型或大小不正确")
	}
	storageContext, cancel := context.WithTimeout(ctx, time.Duration(a.config.Minio.RequestTimeoutMS)*time.Millisecond)
	defer cancel()
	if err := a.ensureBucket(storageContext); err != nil {
		return nil, errors.NewHTTP(503, "媒体存储暂时不可用")
	}
	suffix := unsafeMediaSuffix.ReplaceAllString(strings.ToLower(path.Ext(*input.FileName)), "")
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	now := time.Now().UTC()
	objectKey := now.Format("2006/01/02") + "/" + uuid.NewString() + suffix
	signed, err := a.storage.PresignedPutObject(storageContext, a.config.Minio.Bucket, objectKey, 15*time.Minute)
	if err != nil {
		return nil, errors.NewHTTP(503, "媒体存储暂时不可用")
	}
	uploadURL, err := publicUploadURL(signed, a.config.Media.UploadPublicURL)
	if err != nil {
		return nil, errors.NewHTTP(503, "媒体上传地址配置错误")
	}
	media := schema.Media{
		ID:           strings.ReplaceAll(uuid.NewString(), "-", ""),
		ObjectKey:    objectKey,
		OriginalName: *input.FileName,
		MimeType:     input.MimeType,
		Size:         input.Size,
		Status:       "PENDING",
		UploaderID:   uploaderID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := a.store.Create(ctx, &media); err != nil {
		return nil, err
	}
	return &schema.MediaPresignResult{
		ID:        media.ID,
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresIn: 900,
		Headers:   map[string]string{"Content-Type": input.MimeType},
	}, nil
}

func (a *Media) Complete(ctx context.Context, uploaderID, id string, input schema.MediaCompleteInput) (*schema.MediaCompleteResult, error) {
	if !input.Valid() {
		return nil, errors.NewHTTP(400, "图片尺寸或描述不正确")
	}
	media, err := a.store.FindPending(ctx, id, uploaderID)
	if err != nil {
		return nil, err
	}
	if media == nil {
		return nil, errors.NewHTTP(404, "待确认的媒体不存在")
	}
	storageContext, cancel := context.WithTimeout(ctx, time.Duration(a.config.Minio.RequestTimeoutMS)*time.Millisecond)
	defer cancel()
	stat, err := a.storage.StatObject(storageContext, a.config.Minio.Bucket, media.ObjectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, errors.NewHTTP(400, "对象尚未上传完成")
	}
	if stat.Size != media.Size {
		return nil, errors.NewHTTP(400, "上传文件大小与声明不一致")
	}
	if stat.ContentType != "" && stat.ContentType != media.MimeType {
		return nil, errors.NewHTTP(400, "上传文件类型与声明不一致")
	}
	updated, err := a.store.Complete(ctx, id, uploaderID, input, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, errors.NewHTTP(404, "待确认的媒体不存在")
	}
	return &schema.MediaCompleteResult{
		Media: *updated,
		URL:   strings.TrimRight(a.config.Media.PublicURL, "/") + "/" + media.ObjectKey,
	}, nil
}
