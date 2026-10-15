package minio

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	internalClient *Client
	once           sync.Once
)

// ErrorInvalidName 无效的文件名
var ErrorInvalidName = errors.New("invalid file name")

// NewClient configures the template's storage plugin without network operations.
// Both dependency injection and the legacy file adapter use this one factory.
func NewClient(endpoint, accessKey, secretKey string, secure bool, timeout time.Duration) (*minio.Client, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = timeout
	transport.DialContext = (&net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}).DialContext
	return minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:    secure,
		Transport: transport,
		Region:    "us-east-1",
	})
}

// Init 初始化minio客户端
func Init(addr, accessKey, secretKey string) *Client {
	once.Do(func() {
		cli, err := NewClient(addr, accessKey, secretKey, false, 5*time.Second)
		if err != nil {
			panic(err)
		}
		internalClient = &Client{cli}
	})
	return internalClient
}

// GetClient 获取文件存储客户端
func GetClient() *Client {
	return internalClient
}

// Client minio客户端
type Client struct {
	cli *minio.Client
}

// MinioClient 文件存储客户端
func (a *Client) MinioClient() *minio.Client {
	return a.cli
}

// Store 保存文件
// filename 前3段约束(第一段约束为(前缀)，第二段为bucket(业务类型)，第三段以后为文件key)
func (a *Client) Store(ctx context.Context, filename string, data io.Reader, size int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	bucket, objName, err := a.parseFilename(filename)
	if err != nil {
		return err
	}

	exists, err := a.cli.BucketExists(ctx, bucket)
	if err != nil {
		return err
	} else if !exists {
		err = a.cli.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			return err
		}
	}

	buf, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	rd := bytes.NewBuffer(buf)
	if size == 0 {
		size = int64(rd.Len())
	}

	_, err = a.cli.PutObject(ctx, bucket, objName, rd, size, minio.PutObjectOptions{
		ContentType: http.DetectContentType(buf),
		NumThreads:  2,
	})

	return err
}

// 解析文件名
func (a *Client) parseFilename(filename string) (string, string, error) {
	if len(filename) > 0 && filename[0] == '/' {
		filename = filename[1:]
	}
	names := strings.Split(filename, "/")
	if len(names) < 3 {
		return "", "", ErrorInvalidName
	}

	return strings.ToLower(names[1]), strings.Join(names[2:], "/"), nil
}

// Get 获取文件对象
func (a *Client) Get(ctx context.Context, filename string) (*minio.Object, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	bucketName, objectName, err := a.parseFilename(filename)
	if err != nil {
		return nil, err
	}

	return a.cli.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
}

// Stat 文件状态信息
func (a *Client) Stat(filename string) (minio.ObjectInfo, error) {
	bucketName, objectName, err := a.parseFilename(filename)
	if err != nil {
		return minio.ObjectInfo{}, err
	}

	return a.cli.StatObject(context.Background(), bucketName, objectName, minio.StatObjectOptions{})
}

// GetBase64 获取文件的base64数据
func (a *Client) GetBase64(ctx context.Context, filename string) (string, error) {
	obj, err := a.Get(ctx, filename)
	if err != nil {
		return "", err
	}
	defer obj.Close()

	buf, err := io.ReadAll(obj)
	if err != nil {
		return "", err
	}

	s := base64.StdEncoding.EncodeToString(buf)
	return s, nil
}
