package tests

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	server "github.com/iurikman/smartSurvey/internal/rest"
)

func (s *IntegrationTestSuite) TestCompanies() {
	rCompany := new(models.Company)
	company1 := models.Company{
		ID:   s.testCompanyID,
		Name: "testCompanyPost1",
	}
	company2 := models.Company{
		ID:   uuid.New(),
		Name: "testCompanyPost2",
	}
	company3 := models.Company{
		ID:   uuid.New(),
		Name: "",
	}
	company4 := models.Company{
		ID:   uuid.New(),
		Name: "testCompanyPost4",
	}

	s.Run("POST:", func() {
		s.Run("201/created", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				companiesEndpoint,
				company1,
				&server.HTTPResponse{Data: &rCompany},
			)
			s.Require().Equal(http.StatusCreated, resp.StatusCode)
			s.testCompanyID = rCompany.ID
		})

		s.Run("422/StatusUnprocessableEntity/companyNameIsEmpty", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				companiesEndpoint,
				company3,
				nil,
			)
			s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
		})

		s.Run("409/StatusConflict/duplicate company", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				companiesEndpoint,
				company1,
				&server.HTTPResponse{Data: &rCompany},
			)
			s.Require().Equal(http.StatusConflict, resp.StatusCode)
		})
	})

	s.Run("PATCH:", func() {
		s.Run("404/StatusNotFound/userNotFound", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPatch,
				companiesEndpoint+"/"+uuid.New().String(),
				company2,
				&server.HTTPResponse{Data: &rCompany},
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})

		s.Run("422/StatusUnprocessableEntity/companyNameIsEmpty", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				usersEndpoint,
				company3,
				nil,
			)
			s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
		})
	})

	s.Run("GET:", func() {
		s.Run("list with sorting=name and descending=true", func() {
			var companies []models.Company

			var company models.Company

			params := "?limit=10&sorting=name&descending=true"

			resp1 := s.sendRequest(
				context.Background(),
				http.MethodPost,
				companiesEndpoint,
				company4,
				&server.HTTPResponse{Data: &company})
			s.Require().Equal(http.StatusCreated, resp1.StatusCode)

			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				companiesEndpoint+params,
				nil,
				&server.HTTPResponse{Data: &companies},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(2, len(companies))
			s.Require().Equal(company4.Name, companies[0].Name)
		})
	})
}
