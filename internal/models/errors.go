package models

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateUser      = errors.New("user is already exist")
	ErrNilUUID            = errors.New("uuid id nil")
	ErrNotAllowed         = errors.New("not allowed")
	ErrDuplicateCompany   = errors.New("duplicate company")
	ErrCompanyNotFound    = errors.New("company not found")
	ErrCompanyNameIsEmpty = errors.New("name is empty")
	ErrUserNameIsEmpty    = errors.New("username is empty")
	ErrEmailIsEmpty       = errors.New("email is empty")
	ErrPhoneIsEmpty       = errors.New("phone is empty")
	ErrEmptyRequest       = errors.New("empty request")
	ErrInvalidAccessToken = errors.New("invalid access token")
	ErrHeaderIsEmpty      = errors.New("authHeader is empty")
	ErrBucketIsEmpty      = errors.New("bucket is empty")
	ErrUploadFile         = errors.New("upload file error")
	ErrGetFile            = errors.New("get file error")
)
