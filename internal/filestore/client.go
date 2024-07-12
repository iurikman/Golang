package filestore

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

const (
	ctxTimeout = 60 * time.Second
)

type Object struct {
	ID   uuid.UUID
	Size int64
	Tags map[string]string
}

type Client struct {
	minioClient *minio.Client
	logger      log.Logger
}

func newClient(endpoint, accessKeyID, secretAccessKey string) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &Client{minioClient: minioClient}, nil
}

func (c *Client) uploadFile(ctx context.Context, fileID uuid.UUID, fileName string, bucketName string,
	fileSize int64, reader io.Reader,
) (*models.File, error) {
	var uploadedFileData models.File

	reqCtx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()

	bucketExists, err := c.minioClient.BucketExists(reqCtx, bucketName)
	if err != nil {
		c.logger.Errorf("c.minioClient.BucketExists(reqCtx, bucketName) err: %v", err)
	}

	if !bucketExists {
		c.logger.Warnf("there is no #{bucketName} bucket, creating a new one")

		err := c.minioClient.MakeBucket(reqCtx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			c.logger.Errorf("c.minioClient.MakeBucket(reqCtx, bucketName, minio.MakeBucketOptions{}) err: %v", err)
		}
	}

	uploadedFileInfo, err := c.minioClient.PutObject(
		reqCtx,
		bucketName,
		fileID.String(),
		reader,
		fileSize,
		minio.PutObjectOptions{
			UserMetadata: map[string]string{"Name": fileName, "id": fileID.String(), "BucketName": bucketName},
			ContentType:  "application/octet-stream",
		},
	)
	if err != nil {
		c.logger.Errorf("c.minioClient.PutObject(...) err: %v", err)
	}

	uploadedFileData.ID = fileID

	uploadedFileData.Name = uploadedFileInfo.Bucket

	uploadedFileData.Size = uploadedFileInfo.Size

	uploadedFileData.Bytes = nil

	return &uploadedFileData, nil
}

func (c *Client) getFile(ctx context.Context, bucketName string, fileID uuid.UUID) (*minio.Object, error) {
	obj, err := c.minioClient.GetObject(ctx, bucketName, fileID.String(), minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("c.minioClient.GetObject(...) err: %w", err)
	}

	return obj, nil
}

func (c *Client) GetBucketFiles(ctx context.Context, bucketName string) ([]*minio.Object, error) {
	//nolint:prealloc
	var bucketFiles []*minio.Object

	reqCtx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()

	for obj := range c.minioClient.ListObjects(reqCtx, bucketName, minio.ListObjectsOptions{WithMetadata: true}) {
		if obj.Err != nil {
			c.logger.Errorf(" c.minioClient.ListObjects(...) err: %v", obj.Err)
		}

		object, err := c.minioClient.GetObject(reqCtx, bucketName, obj.Key, minio.GetObjectOptions{})
		if err != nil {
			c.logger.Errorf("c.minioClient.GetObject(reqCtx, bucketName, obj.Key, minio.GetObjectOptions{}) err: %v", err)
		}

		bucketFiles = append(bucketFiles, object)
	}

	return bucketFiles, nil
}

func (c *Client) DeleteFile(ctx context.Context, fileID uuid.UUID, fileName string) error {
	err := c.minioClient.RemoveObject(ctx, fileID.String(), fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("c.minioClient.RemoveObject(reqCtx, bucketName, fileName) err: %w", err)
	}

	return nil
}
