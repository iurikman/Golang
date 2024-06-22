package minio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	"github.com/iurikman/smartSurvey/internal/store"
	log "github.com/sirupsen/logrus"
)

type minioStorage struct {
	client *Client
}

func NewMinioStorage(endpoint, accessKeyID, secretAccessKey string) (store.Storage, error) {
	client, err := NewClient(endpoint, accessKeyID, secretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("NewClient(endpoint, accessKeyID, secretAccessKey) err: %w", err)
	}

	return &minioStorage{
		client: client,
	}, nil
}

func (m *minioStorage) UploadFile(ctx context.Context, file *models.File) (*models.File, error) {
	uploadedFileData, err := m.client.UploadFile(ctx, file.ID, file.Name, file.Name, file.Size,
		bytes.NewBuffer(file.Bytes))
	if err != nil {
		return nil, fmt.Errorf("m.client.UploadFile(ctx, %s, %s, %d) err: %w", file.ID, file.ID.String(), file.Size, err)
	}

	return uploadedFileData, nil
}

func (m *minioStorage) GetFile(ctx context.Context, bucketName string, fileID uuid.UUID) (*models.File, error) {
	obj, err := m.client.GetFile(ctx, bucketName, fileID)
	if err != nil {
		return nil, fmt.Errorf("m.client.GetFile(ctx, bucketName, fileID) err: %w", err)
	}

	defer obj.Close()

	objectInfo, err := obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("obj.Stat() err: %w", err)
	}

	buffer := make([]byte, objectInfo.Size)

	_, err = obj.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("obj.Read() err: %w", err)
	}

	fileID, err = uuid.Parse(objectInfo.Key)
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse() err: %w", err)
	}

	file := models.File{
		ID:    fileID,
		Name:  objectInfo.UserMetadata["Name"],
		Size:  objectInfo.Size,
		Bytes: buffer,
	}

	return &file, nil
}

func (m *minioStorage) GetBucketFiles(ctx context.Context, bucketName string) ([]*models.File, error) {
	var files []*models.File

	objects, err := m.client.GetBucketFiles(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("m.client.GetBucketFiles(ctx, %s) err: %w", bucketName, err)
	}

	if len(objects) == 0 {
		return nil, models.ErrBucketIsEmpty
	}

	for _, obj := range objects {
		stat, err := obj.Stat()
		if err != nil {
			log.Errorf("obj.Stat() err: %v", err)

			continue
		}

		buffer := make([]byte, stat.Size)

		_, err = obj.Read(buffer)
		if err != nil || errors.Is(err, io.EOF) {
			log.Errorf("obj.Read() err: %v", err)

			continue
		}

		id, err := uuid.Parse(stat.UserMetadata["ID"])
		if err != nil {
			log.Warnf("uuid.Parse(\"%s\") err: %v", stat.UserMetadata["ID"], err)
		}

		file := models.File{
			ID:    id,
			Name:  stat.UserMetadata["Name"],
			Size:  stat.Size,
			Bytes: buffer,
		}
		files = append(files, &file)
		_ = obj.Close()
	}

	return files, nil
}

func (m *minioStorage) DeleteFile(ctx context.Context, fileID uuid.UUID, fileName string) error {
	err := m.client.DeleteFile(ctx, fileID, fileName)
	if err != nil {
		return fmt.Errorf("m.client.DeleteFile(ctx, %s, %s) err: %w", fileName, fileName, err)
	}

	return nil
}
