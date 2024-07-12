package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
)

type db interface {
	CreateUser(ctx context.Context, user models.User) (*models.User, error)
	GetUsers(ctx context.Context, params models.GetParams) ([]*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	CreateCompany(ctx context.Context, company models.Company) (*models.Company, error)
	GetCompanies(ctx context.Context, params models.GetParams) ([]*models.Company, error)
	UpdateCompany(ctx context.Context, company models.Company) (*models.Company, error)
	// UploadFile(ctx context.Context, file models.File) (*models.File, error)
	// GetFile(ctx context.Context, file models.File) (*models.File, error)
}

type fileStore interface {
	GetFile(ctx context.Context, bucketName string, fileID uuid.UUID) (*models.File, error)
	GetBucketFiles(ctx context.Context, bucketName string) ([]*models.File, error)
	UploadFile(ctx context.Context, file *models.File) (*models.File, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID, fileName string) error
}

type Service struct {
	db      db
	storage fileStore
}

func New(db db, storage fileStore) *Service {
	return &Service{
		db:      db,
		storage: storage,
	}
}

func (s *Service) UploadFile(ctx context.Context, fileDTO models.FileDTO) (*models.File, error) {
	file := models.NewFile(fileDTO)

	uploadedFileData, err := s.storage.UploadFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("s.storage.UploadFile(ctx, file.ID, bucketName, file) err: %w", err)
	}

	return uploadedFileData, nil
}

func (s *Service) GetFile(ctx context.Context, bucketName string, fileID uuid.UUID) (*models.File, error) {
	file, err := s.storage.GetFile(ctx, bucketName, fileID)
	if err != nil {
		return nil, fmt.Errorf("s.storage.GetFile(ctx, bucketName, fileID) err: %w", err)
	}

	return file, nil
}

func (s *Service) DeleteFile(ctx context.Context, fileID uuid.UUID, fileName string) error {
	//TODO implement func
	return nil
}

func (s *Service) CreateCompany(ctx context.Context, company models.Company) (*models.Company, error) {
	if err := company.Validate(); err != nil {
		return nil, fmt.Errorf("company.Validate(): %w", err)
	}

	rCompany, err := s.db.CreateCompany(ctx, company)
	if err != nil {
		return nil, fmt.Errorf("s.db.CreateCompany(ctx, company) err: %w", err)
	}

	return rCompany, nil
}

func (s *Service) GetCompanies(ctx context.Context, params models.GetParams) ([]*models.Company, error) {
	companies, err := s.db.GetCompanies(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetCompanies(ctx, %v) err: %w", params, err)
	}

	return companies, nil
}

func (s *Service) UpdateCompany(ctx context.Context, company models.Company) (*models.Company, error) {
	if err := company.Validate(); err != nil {
		return nil, fmt.Errorf("company.Validate() err: %w", err)
	}

	rCompany, err := s.db.UpdateCompany(ctx, company)
	if err != nil {
		return nil, fmt.Errorf("s.db.UpdateCompany(ctx, id) err: %w", err)
	}

	return rCompany, nil
}

func (s *Service) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user.Validate() err: %w", err)
	}

	rUser, err := s.db.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("s.db.CreateUser(ctx, user): %w", err)
	}

	return rUser, nil
}

func (s *Service) GetUsers(ctx context.Context, params models.GetParams) ([]*models.User, error) {
	users, err := s.db.GetUsers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetUsers() err: %w", err)
	}

	return users, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.db.GetUserByID(ctx, id): %w", err)
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, user models.UpdateUserRequest) (*models.User, error) {
	newUser, err := s.db.UpdateUser(ctx, id, user)
	if err != nil {
		return nil, fmt.Errorf("s.db.PatchUser(ctx, user): %w", err)
	}

	return newUser, nil
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.db.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("s.db.DeleteUser(ctx, id): %w", err)
	}

	return nil
}
