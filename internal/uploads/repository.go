package uploads

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadRepository interface {
	createFile(db *gorm.DB, model *File) error
	getFileByID(db *gorm.DB, id uuid.UUID) (*File, error)
	getQuotaByUserID(db *gorm.DB, userID uuid.UUID) (*UserQuota, error)
	incrementUsedBytes(db *gorm.DB, userID uuid.UUID, usedBytes int64) error
	upsertQuota(db *gorm.DB, quota *UserQuota) error
}

type uploadRepository struct{}

func NewUploadRepository() UploadRepository {
	return &uploadRepository{}
}

func (r *uploadRepository) createFile(db *gorm.DB, model *File) error {
	return db.Create(model).Error
}

func (r *uploadRepository) getFileByID(db *gorm.DB, id uuid.UUID) (*File, error) {
	var file File
	if err := db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *uploadRepository) getQuotaByUserID(db *gorm.DB, userID uuid.UUID) (*UserQuota, error) {
	var quota UserQuota
	if err := db.First(&quota, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &quota, nil
}

func (r *uploadRepository) incrementUsedBytes(db *gorm.DB, userID uuid.UUID, usedBytes int64) error {
	result := db.Model(&UserQuota{}).
		Where("user_id = ?", userID).
		Update("used_bytes", gorm.Expr("used_bytes + ?", usedBytes))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// No quota record yet — create one with the default max
		quota := &UserQuota{
			UserID:    userID,
			MaxBytes:  DefaultMaxBytes,
			UsedBytes: usedBytes,
		}
		return db.Create(quota).Error
	}
	return nil
}

func (r *uploadRepository) upsertQuota(db *gorm.DB, quota *UserQuota) error {
	return db.Where("user_id = ?", quota.UserID).Assign(quota).FirstOrCreate(quota).Error
}
