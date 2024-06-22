package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/iurikman/smartSurvey/internal/models"
	log "github.com/sirupsen/logrus"
)

const (
	StandardPage int = 10
)

type HTTPResponse struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

type service interface { //nolint:interfacebloat
	CreateUser(ctx context.Context, user models.User) (*models.User, error)
	GetUsers(ctx context.Context, params models.GetParams) ([]*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	CreateCompany(ctx context.Context, company models.Company) (*models.Company, error)
	GetCompanies(ctx context.Context, params models.GetParams) ([]*models.Company, error)
	UpdateCompany(ctx context.Context, company models.Company) (*models.Company, error)
	UploadFile(ctx context.Context, file models.FileDTO) (*models.File, error)
	GetFile(ctx context.Context, fileName string, fileID uuid.UUID) (*models.File, error)
	GetBucketFiles(ctx context.Context, bucketName string) ([]*models.File, error)
	DeleteFile(ctx context.Context, bucketName, fileName string) error
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/form-data; boundary=AaB03x")

	var file models.File

	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	fileDTO := models.FileDTO{
		Name:   file.Name,
		Size:   file.Size,
		Reader: bytes.NewReader(file.Bytes),
	}

	uploadedFileData, err := s.service.UploadFile(r.Context(), fileDTO)
	if err != nil {
		log.Warnf("s.service.UploadFile(r.Context(), fileDTO) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusCreated, uploadedFileData)
}

func (s *Server) getFile(w http.ResponseWriter, r *http.Request) {
	bucketName := r.URL.Query().Get("bucketName")
	fileID := chi.URLParam(r, "id")

	id, err := uuid.Parse(fileID)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
	}

	file, err := s.service.GetFile(r.Context(), bucketName, id)
	if err != nil {
		log.Warnf("s.service.GetFile(r.Context(), bucketName, id) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, file)
}

func (s *Server) getBucketFiles(w http.ResponseWriter, r *http.Request) {
	bucketName := chi.URLParam(r, "bucketName")

	files, err := s.service.GetBucketFiles(r.Context(), bucketName)
	if err != nil {
		log.Warnf("s.service.GetBucketFiles(r.Context(), bucketName) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, files)
}

func (s *Server) deleteFile(w http.ResponseWriter, r *http.Request) {
	bucketName := r.URL.Query().Get("bucketName")
	fileID := chi.URLParam(r, "id")

	err := s.service.DeleteFile(r.Context(), bucketName, fileID)
	if err != nil {
		log.Warnf("s.service.DeleteFile(r.Context(), bucketName, fileID) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createCompany(w http.ResponseWriter, r *http.Request) {
	var rCompany models.Company

	userInfo, ok := r.Context().Value(models.UserInfoKey).(models.UserInfo)
	if !ok {
		log.Warn("User info not found in context")
	}

	if err := json.NewDecoder(r.Body).Decode(&rCompany); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
	}

	rCompany.ID = userInfo.ID

	company, err := s.service.CreateCompany(r.Context(), rCompany)

	switch {
	case errors.Is(err, models.ErrCompanyNameIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrDuplicateCompany):
		writeErrorResponse(w, http.StatusConflict, err.Error())

		return
	case err != nil:
		log.Warnf("s.service.CreateCompany(r.Context(), rCompany) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusCreated, company)
}

func (s *Server) getCompanies(w http.ResponseWriter, r *http.Request) {
	params, err := ParseParams(r.URL.Query())
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	companies, err := s.service.GetCompanies(r.Context(), *params)
	if err != nil {
		log.Warnf("s.service.GetCompanies(r.Context(), *params) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, companies)
}

func (s *Server) updateCompany(w http.ResponseWriter, r *http.Request) {
	var rCompany models.Company

	if err := json.NewDecoder(r.Body).Decode(&rCompany); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
	}

	company, err := s.service.UpdateCompany(r.Context(), rCompany)

	switch {
	case errors.Is(err, models.ErrCompanyNameIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrCompanyNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		log.Warnf("s.service.UpdateCompany(r.Context(), rCompany) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, company)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var rUser models.User

	if err := json.NewDecoder(r.Body).Decode(&rUser); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	user, err := s.service.CreateUser(r.Context(), rUser)

	switch {
	case
		errors.Is(err, models.ErrUserNameIsEmpty) ||
			errors.Is(err, models.ErrEmailIsEmpty) ||
			errors.Is(err, models.ErrPhoneIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrDuplicateUser):
		writeErrorResponse(w, http.StatusConflict, err.Error())

		return
	case errors.Is(err, models.ErrCompanyNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		log.Warnf("s.service.CreateUser(r.Context(), rUser) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusCreated, user)
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	params, err := ParseParams(r.URL.Query())
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	users, err := s.service.GetUsers(r.Context(), *params)
	if err != nil {
		log.Warnf("s.service.GetUsers(r.Context(), *params) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, users)
}

func (s *Server) getUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	user, err := s.service.GetUserByID(r.Context(), id)

	switch {
	case errors.Is(err, models.ErrUserNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		log.Warnf("s.service.GetUserByID(r.Context(), id) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, user)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	var updateRequest models.UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	user, err := s.service.UpdateUser(r.Context(), id, updateRequest)

	switch {
	case errors.Is(err, models.ErrUserNameIsEmpty) || errors.Is(err, models.ErrEmailIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrUserNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case errors.Is(err, models.ErrDuplicateUser):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrNotAllowed):
		writeErrorResponse(w, http.StatusUnauthorized, err.Error())

		return
	case errors.Is(err, models.ErrEmptyRequest):
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	case err != nil:
		log.Warnf("s.service.PatchUser(r.Context(), id, patchRequest) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, user)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
	}

	err = s.service.DeleteUser(r.Context(), id)

	switch {
	case errors.Is(err, models.ErrUserNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		log.Warnf("s.service.DeleteUser(r.Context(), id) err: %v", err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeOkResponse(w http.ResponseWriter, statusCode int, respData any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(HTTPResponse{Data: respData})
	if err != nil {
		log.Warn("writeOkResponse/json.NewEncoder(w).Encode(HTTPResponse{Data: data})")
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(HTTPResponse{Error: description})
	if err != nil {
		log.Warn("writeErrorResponse/json.NewEncoder(w).Encode(HTTPResponse{Data: data})")
	}
}

func ParseParams(value url.Values) (*models.GetParams, error) {
	decoder := schema.NewDecoder()

	params := &models.GetParams{}

	err := decoder.Decode(params, value)
	if err != nil {
		return nil, fmt.Errorf("decoder.Decode(params, value) err: %w", err)
	}

	if params.Limit == 0 {
		params.Limit = StandardPage
	}

	return params, nil
}
