package uploads

import (
	"context"

	"github.com/alkuinvito/ai-assistant/pkg/storage"
)

type UploadService interface {
	CreatePresignedPutUploadUrl(ctx context.Context, bucket, path string, fileSize int64, exp ...int) (DTOPresignedPutUrl, error)
	CreateUploadUrl(ctx context.Context, bucket, path string, exp ...int) (DTOUploadUrl, error)
}

type uploadService struct {
	storage storage.Storage
}

func NewUploadService(s storage.Storage) UploadService {
	return &uploadService{s}
}

func (s *uploadService) CreatePresignedPutUploadUrl(ctx context.Context, bucket, path string, fileSize int64, exp ...int) (DTOPresignedPutUrl, error) {
	url, err := s.storage.PresignedPutUpload(ctx, bucket, path, fileSize, exp...)
	if err != nil {
		return DTOPresignedPutUrl{}, err
	}

	return DTOPresignedPutUrl{Url: url}, nil
}

func (s *uploadService) CreateUploadUrl(ctx context.Context, bucket, path string, exp ...int) (DTOUploadUrl, error) {
	endpoint, formData, err := s.storage.PresignedPostUpload(ctx, bucket, path, exp...)
	if err != nil {
		return DTOUploadUrl{}, err
	}

	return DTOUploadUrl{
		Endpoint: endpoint,
		Fields:   formData,
	}, nil
}
