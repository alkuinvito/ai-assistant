package uploads

import (
	"context"
	"fmt"

	"github.com/alkuinvito/ai-assistant/pkg/service_error"
	"github.com/alkuinvito/ai-assistant/pkg/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadService interface {
	CreatePresignedPutUploadUrl(ctx context.Context, bucket, path string, fileSize int64, userId uuid.UUID, exp ...int) (DTOPresignedPutUrl, error)
	CreateUploadUrl(ctx context.Context, bucket, path string, exp ...int) (DTOUploadUrl, error)
	GetFileByID(db *gorm.DB, fileId uuid.UUID) (*File, error)
	GetFileByIDAndUser(ctx context.Context, db *gorm.DB, fileId uuid.UUID, userId uuid.UUID) (*File, error)
	IncrementUsedBytes(ctx context.Context, userID uuid.UUID, usedBytes int64) error
}

type uploadService struct {
	db      *gorm.DB
	repo    UploadRepository
	storage storage.Storage
}

func NewUploadService(db *gorm.DB, repo UploadRepository, s storage.Storage) UploadService {
	return &uploadService{db, repo, s}
}

func (s *uploadService) CreatePresignedPutUploadUrl(ctx context.Context, bucket, path string, fileSize int64, userId uuid.UUID, exp ...int) (DTOPresignedPutUrl, error) {
	db := s.db.WithContext(ctx)

	// Check user quota
	quota, err := s.repo.getQuotaByUserID(db, userId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return DTOPresignedPutUrl{}, service_error.NewInternalServerError(ctx, "failed to check upload quota", err)
		}
		// First upload — no quota record yet, allow
		quota = nil
	}

	if quota != nil && quota.UsedBytes+fileSize > quota.MaxBytes {
		return DTOPresignedPutUrl{}, service_error.NewForbidden(ctx, "upload quota exceeded. please free up space or contact support", nil)
	}

	// Build a namespaced object path to prevent collisions
	fileId := uuid.New()
	objectPath := fmt.Sprintf("%s/%s", userId.String(), path)

	// Generate presigned PUT URL with Content-Length signed
	url, err := s.storage.PresignedPutUpload(ctx, bucket, objectPath, fileSize, exp...)
	if err != nil {
		return DTOPresignedPutUrl{}, service_error.NewInternalServerError(ctx, "failed to generate presigned upload url", err)
	}

	// Create a File record so the file is tracked before it's saved as an invoice
	file := &File{
		Id:       fileId,
		UserId:   userId,
		Filename: path,
		Path:     objectPath,
		Size:     float64(fileSize),
	}
	if err := s.repo.createFile(db, file); err != nil {
		return DTOPresignedPutUrl{}, service_error.NewInternalServerError(ctx, "failed to create file record", err)
	}

	return DTOPresignedPutUrl{Url: url, FileId: fileId}, nil
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

func (s *uploadService) IncrementUsedBytes(ctx context.Context, userID uuid.UUID, usedBytes int64) error {
	db := s.db.WithContext(ctx)
	return s.repo.incrementUsedBytes(db, userID, usedBytes)
}
