package models

import (
	"fmt"
	"io"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
)

const (
	StandardPage int    = 10
	UserInfoKey  ctxKey = "userInfo"
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

// TODO remove nilable fields
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

// TODO move to rest package
func ParseParams(value url.Values) (*GetParams, error) {
	decoder := schema.NewDecoder()

	params := &GetParams{}

	err := decoder.Decode(params, value)
	if err != nil {
		return nil, fmt.Errorf("decoder.Decode(params, value) err: %w", err)
	}

	if params.Limit == 0 {
		params.Limit = StandardPage
	}

	return params, nil
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
