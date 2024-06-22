package models

import (
	"io"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	UserInfoKey ctxKey = "userInfo"
)

type File struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Size  int64     `json:"size"`
	Bytes []byte    `json:"bytes"`
}

type FileDTO struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Reader io.Reader
}

func NewFile(dto FileDTO) *File {
	buffer, _ := io.ReadAll(dto.Reader)
	file := File{
		ID:    uuid.New(),
		Name:  dto.Name,
		Size:  dto.Size,
		Bytes: buffer,
	}

	return &file
}

type UpdateUserRequest struct {
	Company  uuid.UUID `json:"company,omitempty"`
	Role     *string   `json:"role,omitempty"`
	Name     *string   `json:"name"`
	Surname  *string   `json:"surname,omitempty"`
	Phone    *string   `json:"phone"`
	Email    *string   `json:"email"`
	UserType *string   `json:"userType,omitempty"`
}

type User struct {
	ID       uuid.UUID `json:"id"`
	Company  uuid.UUID `json:"company,omitempty"`
	Role     *string   `json:"role,omitempty"`
	Name     *string   `json:"name"`
	Surname  *string   `json:"surname,omitempty"`
	Phone    *string   `json:"phone"`
	Email    *string   `json:"email"`
	UserType *string   `json:"userType,omitempty"`
}

type ctxKey string

type UserInfo struct {
	ID uuid.UUID
}

type Company struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type GetParams struct {
	Offset     int    `schema:"offset"`
	Limit      int    `schema:"limit"`
	Sorting    string `schema:"sorting"`
	Descending bool   `schema:"descending"`
	Filter     string `schema:"filter"`
}

type Claims struct {
	jwt.RegisteredClaims
	UUID uuid.UUID `json:"uuid"`
}

func (c Company) Validate() error {
	if c.Name == "" {
		return ErrCompanyNameIsEmpty
	}

	return nil
}

func (u User) Validate() error {
	if *u.Name == "" {
		return ErrUserNameIsEmpty
	}

	if *u.Email == "" {
		return ErrEmailIsEmpty
	}

	if *u.Phone == "" {
		return ErrPhoneIsEmpty
	}

	return nil
}

func (u UpdateUserRequest) Validate() error {
	if *u.Name == "" {
		return ErrUserNameIsEmpty
	}

	if *u.Email == "" {
		return ErrEmailIsEmpty
	}

	if *u.Phone == "" {
		return ErrPhoneIsEmpty
	}

	return nil
}
