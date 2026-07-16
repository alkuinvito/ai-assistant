package uploads

import (
	"context"

	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/alkuinvito/ai-assistant/pkg/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadService interface {
	CreatePresignedPutUploadUrl(ctx context.Context, bucket, path string, fileSize int64, exp ...int) (DTOPresignedPutUrl, error)
	CreateUploadUrl(ctx context.Context, bucket, path string, exp ...int) (DTOUploadUrl, error)
	GetFileByID(db *gorm.DB, fileId uuid.UUID) (*File, error)
	GetFileByIDAndUser(ctx context.Context, db *gorm.DB, fileId uuid.UUID, userId uuid.UUID) (*File, error)
}

type uploadService struct {
	db      *gorm.DB
	repo    UploadRepository
	storage storage.Storage
}

func NewUploadService(db *gorm.DB, repo UploadRepository, s storage.Storage) UploadService {
	return &uploadService{db, repo, s}
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

func (s *uploadService) GetFileByID(db *gorm.DB, fileId uuid.UUID) (*File, error) {
	return s.repo.getFileByID(db, fileId)
}

func (s *uploadService) GetFileByIDAndUser(ctx context.Context, db *gorm.DB, fileId uuid.UUID, userId uuid.UUID) (*File, error) {
	file, err := s.repo.getFileByID(db, fileId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, service_error.NewNotFound(ctx, "file not found", err)
		}
		return nil, service_error.NewInternalServerError(ctx, "failed to retrieve file", err)
	}

	if file.UserId != userId {
		return nil, service_error.NewForbidden(ctx, "file does not belong to user", nil)
	}

	return file, nil
}
