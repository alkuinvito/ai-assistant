package uploads

import "github.com/google/uuid"

const DefaultMaxBytes int64 = 512 * 1024 * 1024 // 512 MB per user

type DTOCreatePresignedPutUrl struct {
	Filename string `json:"filename" validate:"required"`
	Size     int64  `json:"size" validate:"required,min=1,max=10485760"`
}

type DTOCreateUploadUrl struct {
	Filename string `json:"filename" validate:"required"`
}

type DTOPresignedPutUrl struct {
	Url    string    `json:"url"`
	FileId uuid.UUID `json:"file_id"`
}

type DTOUploadFileResponse struct {
	Id       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
}

type DTOUploadUrl struct {
	Endpoint string            `json:"endpoint"`
	Fields   map[string]string `json:"fields"`
}
