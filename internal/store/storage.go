package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
)

type Storage interface {
	GetFile(ctx context.Context, bucketName string, fileID uuid.UUID) (*models.File, error)
	GetBucketFiles(ctx context.Context, bucketName string) ([]*models.File, error)
	UploadFile(ctx context.Context, file *models.File) (*models.File, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID, fileName string) error
}
