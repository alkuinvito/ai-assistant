package storage

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	GetObject(ctx context.Context, bucketName, objectName string) ([]byte, error)
	PresignedPostUpload(ctx context.Context, bucketName, objectName string, expiresIn ...int) (string, map[string]string, error)
	PresignedPutUpload(ctx context.Context, bucketName, objectName string, fileSize int64, expiresIn ...int) (string, error)
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error
	RemoveObject(ctx context.Context, bucketName, objectName string) error
}

type minioStorage struct {
	client *minio.Client
}

func NewMinioStorage(logger *logrus.Logger) (Storage, error) {
	client, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
		Secure: true,
		Region: "auto",
	})
	if err != nil {
		return nil, err
	}

	return &minioStorage{client}, nil
}

func (m *minioStorage) GetObject(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	reader, err := m.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (m *minioStorage) PresignedPostUpload(ctx context.Context, bucketName, objectName string, expiresIn ...int) (string, map[string]string, error) {
	exp := 0
	for _, t := range expiresIn {
		exp += t
	}

	policy := minio.NewPostPolicy()

	policy.SetBucket(bucketName)
	policy.SetKey(objectName)

	if exp != 0 {
		policy.SetExpires(time.Now().UTC().Add(time.Duration(exp) * time.Second))
	}

	// Max 10MB (min 1 byte to avoid empty uploads)
	if err := policy.SetContentLengthRange(1, 10*1024*1024); err != nil {
		return "", nil, err
	}

	url, formData, err := m.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return "", nil, err
	}

	return url.String(), formData, nil
}

func (m *minioStorage) PresignedPutUpload(ctx context.Context, bucketName, objectName string, fileSize int64, expiresIn ...int) (string, error) {
	expiry := time.Hour
	for _, t := range expiresIn {
		expiry = time.Duration(t) * time.Second
		break
	}

	headers := make(http.Header)
	headers.Set("Content-Length", strconv.FormatInt(fileSize, 10))

	u, err := m.client.PresignHeader(ctx, "PUT", bucketName, objectName, expiry, url.Values{}, headers)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (m *minioStorage) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *minioStorage) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	return m.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}
