package uploads

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadRepository interface {
	createFile(db *gorm.DB, model *File) error
	getFileByID(db *gorm.DB, id uuid.UUID) (*File, error)
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
