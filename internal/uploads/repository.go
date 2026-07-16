package uploads

import "gorm.io/gorm"

type UploadRepository interface {
	createFile(db *gorm.DB, model *File) error
}

type uploadRepository struct{}

func NewUploadRepository() UploadRepository {
	return &uploadRepository{}
}

func (r *uploadRepository) createFile(db *gorm.DB, model *File) error {
	return db.Create(model).Error
}
