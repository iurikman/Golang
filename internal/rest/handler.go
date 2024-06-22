package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	log "github.com/sirupsen/logrus"
)

type HTTPResponse struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

type TransferResponse struct {
	TransactionID uuid.UUID `json:"transactionId"`
}

type service interface {
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
	DeleteFile(ctx context.Context, fileID uuid.UUID, fileName string) error
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/form-data; boundary=AaB03x")

	var file models.File

	_ = json.NewDecoder(r.Body).Decode(&file)

	fileDTO := models.FileDTO{
		Name:   file.Name,
		Size:   file.Size,
		Reader: bytes.NewReader(file.Bytes),
	}

	uploadedFileData, err := s.service.UploadFile(r.Context(), fileDTO)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	writeOkResponse(w, http.StatusCreated, uploadedFileData)
}

func (s *Server) getFile(w http.ResponseWriter, r *http.Request) {
	bucketName := r.URL.Query().Get("bucketname")
	fileID := chi.URLParam(r, "id")

	id, err := uuid.Parse(fileID)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
	}

	file, err := s.service.GetFile(r.Context(), bucketName, id)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "s.service.GetFile(r.Context(), bucketName, id) err: %w")

		return
	}

	writeOkResponse(w, http.StatusOK, file)
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
		writeErrorResponse(w, http.StatusInternalServerError, "s.service.CreateCompany(r.Context(), rCompany)")

		return
	}

	writeOkResponse(w, http.StatusCreated, company)
}

func (s *Server) getCompanies(w http.ResponseWriter, r *http.Request) {
	params, err := models.ParseParams(r.URL.Query())
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	companies, err := s.service.GetCompanies(r.Context(), *params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "s.service.GetCompanies(r.Context(), params)")

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
		writeErrorResponse(w, http.StatusInternalServerError, "s.service.updateCompany(r.Context(), rCompany)")

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
	case errors.Is(err, models.ErrUserNameIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrEmailIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrPhoneIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrDuplicateUser):
		writeErrorResponse(w, http.StatusConflict, "user already exists")

		return
	case errors.Is(err, models.ErrCompanyNotFound):
		writeErrorResponse(w, http.StatusNotFound, err.Error())

		return
	case err != nil:
		log.Warn("s.service.CreateUser(r.Context(), rUser) err")
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusCreated, user)
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	params, err := models.ParseParams(r.URL.Query())
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())

		return
	}

	users, err := s.service.GetUsers(r.Context(), *params)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "error getting users")

		return
	}

	writeOkResponse(w, http.StatusOK, users)
}

func (s *Server) getUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "can`t parse ID")

		return
	}

	user, err := s.service.GetUserByID(r.Context(), id)

	switch {
	case errors.Is(err, models.ErrUserNotFound):
		writeErrorResponse(w, http.StatusNotFound, "user not found")

		return
	case err != nil:
		log.Warn("s.service.GetUserByID(r.Context(), id) err")
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, user)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "uuid.Parse(chi.URLParam(r, 'id')) err")

		return
	}

	var updateRequest models.UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "json.NewDecoder(r.Body).Decode(&patchRequest) err")
	}

	user, err := s.service.UpdateUser(r.Context(), id, updateRequest)

	switch {
	case errors.Is(err, models.ErrUserNameIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrEmailIsEmpty):
		writeErrorResponse(w, http.StatusUnprocessableEntity, err.Error())

		return
	case errors.Is(err, models.ErrUserNotFound):
		writeErrorResponse(w, http.StatusNotFound, "user not found")

		return
	case errors.Is(err, models.ErrDuplicateUser):
		writeErrorResponse(w, http.StatusUnprocessableEntity, "duplicate user")

		return
	case errors.Is(err, models.ErrNotAllowed):
		writeErrorResponse(w, http.StatusUnauthorized, "not allowed")

		return
	case errors.Is(err, models.ErrEmptyRequest):
		writeErrorResponse(w, http.StatusBadRequest, "empty request")
	case err != nil:
		log.Warn("s.service.PatchUser(r.Context(), id, patchRequest) err")
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error")

		return
	}

	writeOkResponse(w, http.StatusOK, user)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "uuid.Parse(chi.URLParam(r, 'id')) err")
	}

	err = s.service.DeleteUser(r.Context(), id)

	switch {
	case errors.Is(err, models.ErrUserNotFound):
		writeErrorResponse(w, http.StatusNotFound, "user not found")

		return
	case err != nil:
		log.Warn("s.service.DeleteUser(r.Context(), id) err")
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
