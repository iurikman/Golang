package filestore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	log "github.com/sirupsen/logrus"
)

type MinioStorage struct {
	client *Client
}

func NewMinioStorage(endpoint, accessKeyID, secretAccessKey string) (*MinioStorage, error) {
	client, err := newClient(endpoint, accessKeyID, secretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("newClient(endpoint, accessKeyID, secretAccessKey) err: %w", err)
	}

	return &MinioStorage{
		client: client,
	}, nil
}

func (m *MinioStorage) UploadFile(ctx context.Context, file *models.File) (*models.File, error) {
	uploadedFileData, err := m.client.uploadFile(ctx, file.ID, file.Name, file.Name, file.Size,
		bytes.NewBuffer(file.Bytes))
	if err != nil {
		return nil, fmt.Errorf("m.client.uploadFile(ctx, %s, %s, %d) err: %w", file.ID, file.ID.String(), file.Size, err)
	}

	return uploadedFileData, nil
}

func (m *MinioStorage) GetFile(ctx context.Context, bucketName string, fileID uuid.UUID) (*models.File, error) {
	obj, err := m.client.getFile(ctx, bucketName, fileID)
	if err != nil {
		return nil, fmt.Errorf("m.client.getFile(ctx, bucketName, fileID) err: %w", err)
	}

	defer func() {
		err := obj.Close()
		if err != nil {
			log.Warn("obj.Close() err:", err)
		}
	}()

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

func (m *MinioStorage) GetBucketFiles(ctx context.Context, bucketName string) ([]*models.File, error) {
	objects, err := m.client.getBucketFiles(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("m.client.getBucketFiles(ctx, %s) err: %w", bucketName, err)
	}

	if len(objects) == 0 {
		return nil, models.ErrBucketIsEmpty
	}

	files := make([]*models.File, len(objects))

	for _, obj := range objects {
		stat, err := obj.Stat()
		if err != nil {
			return nil, fmt.Errorf("obj.Stat() err: %w", err)
		}

		buffer := make([]byte, stat.Size)

		_, err = obj.Read(buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("obj.Read() err: %w", err)
		}

		id, err := uuid.Parse(stat.Key)
		if err != nil {
			return nil, fmt.Errorf("uuid.Parse() err: %w", err)
		}

		file := models.File{
			ID:    id,
			Name:  stat.UserMetadata["Name"],
			Size:  stat.Size,
			Bytes: buffer,
		}
		//nolint:makezero
		files = append(files, &file)
		_ = obj.Close()
	}

	return files, nil
}

func (m *MinioStorage) DeleteFile(ctx context.Context, bucketName string, fileName string) error {
	err := m.client.deleteFile(ctx, bucketName, fileName)
	if err != nil {
		return fmt.Errorf("m.client.deleteFile(ctx, %s, %s) err: %w", fileName, fileName, err)
	}

	return nil
}