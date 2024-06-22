package tests

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	server "github.com/iurikman/smartSurvey/internal/rest"
)

func (s *IntegrationTestSuite) checkUserPost(user *models.User) {
	respUserData := new(models.User)

	resp := s.sendRequest(
		context.Background(),
		http.MethodPost,
		usersEndpoint,
		user,
		&server.HTTPResponse{Data: &respUserData},
	)

	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	s.Require().Equal(user.Company, respUserData.Company)
	s.Require().Equal(user.Role, respUserData.Role)
	s.Require().Equal(user.Name, respUserData.Name)
	s.Require().Equal(user.Surname, respUserData.Surname)
	s.Require().Equal(user.Phone, respUserData.Phone)
	s.Require().Equal(user.Email, respUserData.Email)
	s.Require().Equal(user.UserType, respUserData.UserType)
	s.Require().NotZero(respUserData.ID)
	user.ID = respUserData.ID
}

func (s *IntegrationTestSuite) TestUsers() {
	company := models.Company{
		ID:   s.testCompanyID,
		Name: "testCompanyPost",
	}

	rCompany := new(models.Company)
	s.sendRequest(
		context.Background(),
		http.MethodPost,
		companiesEndpoint,
		company,
		&server.HTTPResponse{Data: &rCompany},
	)

	user1 := models.User{
		ID:       uuid.New(),
		Company:  s.testCompanyID,
		Role:     newString("TestRole1"),
		Name:     newString("TestName1"),
		Surname:  newString("TestSurname1"),
		Phone:    newString("1991"),
		Email:    newString("testuser1@test.org"),
		UserType: newString("Type1"),
	}

	user2 := models.User{
		ID:       uuid.New(),
		Company:  s.testCompanyID,
		Role:     newString("TestRole2"),
		Name:     newString("TestName2"),
		Surname:  newString("TestSurname2"),
		Phone:    newString("2991"),
		Email:    newString("testuser2@test.org"),
		UserType: newString("Type2"),
	}

	user3 := models.User{
		ID:       uuid.New(),
		Company:  s.testCompanyID,
		Role:     newString("TestRole3"),
		Name:     newString("TestName3"),
		Surname:  newString("TestSurname3"),
		Phone:    newString("3991"),
		Email:    newString("testuser3@test.org"),
		UserType: newString("Type3"),
	}

	user4 := models.User{
		ID:       uuid.New(),
		Company:  s.testCompanyID,
		Role:     newString("TestRole4"),
		Name:     newString(""),
		Surname:  newString("TestSurname4"),
		Phone:    newString("4991"),
		Email:    newString("testuser4@test.org"),
		UserType: newString("Type4"),
	}

	user5 := models.User{
		ID:       uuid.New(),
		Company:  s.testCompanyID,
		Role:     newString("TestRole5"),
		Name:     newString("TestName5"),
		Surname:  newString("TestSurname5"),
		Phone:    newString("5991"),
		Email:    newString(""),
		UserType: newString("Type5"),
	}

	user6 := models.User{
		ID:       uuid.New(),
		Company:  s.testCompanyID,
		Role:     newString("TestRole6"),
		Name:     newString("TestName6"),
		Surname:  newString("TestSurname6"),
		Phone:    newString(""),
		Email:    newString("testuser6@test.org"),
		UserType: newString("Type6"),
	}

	user7 := models.User{
		ID:       uuid.New(),
		Company:  uuid.New(),
		Role:     newString("TestRole7"),
		Name:     newString("TestName7"),
		Surname:  newString("TestSurname7"),
		Phone:    newString("7991"),
		Email:    newString("testuser7@test.org"),
		UserType: newString("Type7"),
	}

	s.Run("401/StatusUnauthorized", func() {
		temp := s.authToken
		s.authToken = ""

		resp := s.sendRequest(
			context.Background(),
			http.MethodGet,
			usersEndpoint,
			nil,
			nil)
		s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
		s.authToken = temp
	})

	s.Run("POST:", func() {
		s.Run("201/created", func() {
			s.checkUserPost(&user1)
			s.checkUserPost(&user2)
			s.checkUserPost(&user3)
			s.testUserID = user3.ID
		})

		s.Run("409/StatusConflict/duplicate user", func() {
			respUserData := new(models.User)

			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				usersEndpoint,
				user1,
				&server.HTTPResponse{Data: &respUserData},
			)
			s.Require().Equal(http.StatusConflict, resp.StatusCode)
		})

		s.Run("400/badRequest", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				usersEndpoint,
				"badRequest?",
				nil,
			)
			s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
		})
		s.Run("422/StatusUnprocessableEntity/nameIsEmpty", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				usersEndpoint,
				user4,
				nil,
			)
			s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
		})

		s.Run("422/StatusUnprocessableEntity/emailIsEmpty", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				usersEndpoint,
				user5,
				nil,
			)
			s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
		})

		s.Run("422/StatusUnprocessableEntity/phoneIsEmpty", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				usersEndpoint,
				user6,
				nil,
			)
			s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
		})

		s.Run("404/StatusNotFound/companyNotFound", func() {
			resp := s.sendRequest(
				context.Background(),
				http.MethodPost,
				usersEndpoint,
				user7,
				nil,
			)
			s.Require().Equal(http.StatusNotFound, resp.StatusCode)
		})
	})

	s.Run("GET:", func() {
		s.Run("{id}", func() {
			respUserData := new(models.User)
			resp := s.sendRequest(
				context.Background(),
				http.MethodGet,
				usersEndpoint+"/"+user3.ID.String(),
				nil,
				&server.HTTPResponse{Data: &respUserData},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(user3.ID, respUserData.ID)
			s.Require().Equal(user3.Company, respUserData.Company)
			s.Require().Equal(user3.Role, respUserData.Role)
			s.Require().Equal(user3.Name, respUserData.Name)
			s.Require().Equal(user3.Surname, respUserData.Surname)
			s.Require().Equal(user3.Phone, respUserData.Phone)
			s.Require().Equal(user3.Email, respUserData.Email)
			s.Require().Equal(user3.UserType, respUserData.UserType)
		})

		s.Run("usersList", func() {
			s.Run("with sorting=created at and descending=true", func() {
				var respUserData []models.User

				params := "?limit=10&sorting=created_at&descending=true"

				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					usersEndpoint+params,
					nil,
					&server.HTTPResponse{Data: &respUserData},
				)
				s.Require().Equal(http.StatusOK, resp.StatusCode)
				s.Require().Equal(user1.Name, respUserData[2].Name)
				s.Require().Equal(3, len(respUserData))
			})

			limit := 10

			offset := 0

			s.Run("creating 68 users and listing by 10 users descending=true", func() {
				var respUsersData []models.User

				params := fmt.Sprintf("?limit=%d&offset=%d&sorting=created_at&descending=true", limit, offset)

				for i := 1; i < 69; i++ {
					s.checkUserPost(&models.User{
						ID:       uuid.UUID{},
						Company:  s.testCompanyID,
						Role:     newString(strconv.Itoa(i)),
						Name:     newString(strconv.Itoa(i)),
						Surname:  newString(strconv.Itoa(i)),
						Phone:    newString(strconv.Itoa(i)),
						Email:    newString(strconv.Itoa(i)),
						UserType: newString(strconv.Itoa(i)),
					})
				}
				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					usersEndpoint+"/"+params,
					nil,
					&server.HTTPResponse{Data: &respUsersData},
				)
				s.Require().Equal(http.StatusOK, resp.StatusCode)
				s.Require().Equal(10, len(respUsersData))
				s.Require().Equal(newString(strconv.Itoa(68)), respUsersData[0].Name)
			})

			limit = 10
			offset = 10

			s.Run("page 2 users from 8 to 17", func() {
				var respUsersData []models.User

				params := fmt.Sprintf("?limit=%d&offset=%d", limit, offset)
				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					usersEndpoint+"/"+params,
					nil,
					&server.HTTPResponse{Data: &respUsersData},
				)
				s.Require().Equal(http.StatusOK, resp.StatusCode)
				s.Require().Equal(10, len(respUsersData))
				s.Require().Equal(newString(strconv.Itoa(8)), respUsersData[0].Surname)
			})

			limit = 10
			offset = 0
			filter := "1"

			s.Run("filter by name with 1", func() {
				var respUsersData []models.User

				params := fmt.Sprintf("?limit=%d&offset=%d&filter=%s", limit, offset, filter)
				resp := s.sendRequest(
					context.Background(),
					http.MethodGet,
					usersEndpoint+"/"+params,
					nil,
					&server.HTTPResponse{Data: &respUsersData},
				)
				s.Require().Equal(http.StatusOK, resp.StatusCode)
				s.Require().Equal(10, len(respUsersData))
				s.Require().Equal("121314", *respUsersData[4].Name+*respUsersData[5].Name+*respUsersData[6].Name)
			})
		})

		s.Run("PATCH:", func() {
			s.Run("200/StatusOk/newEmail", func() {
				respUserData := new(models.User)

				newName := newString("newEmail")

				resp := s.sendRequest(
					context.Background(),
					http.MethodPatch,
					usersEndpoint+"/"+user1.ID.String(),
					models.UpdateUserRequest{Email: newName},
					&server.HTTPResponse{Data: &respUserData},
				)
				s.Require().Equal(http.StatusOK, resp.StatusCode)
				s.Require().Equal(newName, respUserData.Email)
			})

			s.Run("200/StatusOk/newName newRole", func() {
				respUserData := new(models.User)

				newName := newString("newName2")
				newRole := newString("newRole")

				resp := s.sendRequest(
					context.Background(),
					http.MethodPatch,
					usersEndpoint+"/"+user2.ID.String(),
					models.UpdateUserRequest{Name: newName, Role: newRole},
					&server.HTTPResponse{Data: &respUserData},
				)
				s.Require().Equal(http.StatusOK, resp.StatusCode)
				s.Require().Equal(newRole, respUserData.Role)
				s.Require().Equal(newName, respUserData.Name)
			})

			s.Run("400/StatusBadRequest/emptyRequest", func() {
				respUserData := new(models.User)

				resp := s.sendRequest(
					context.Background(),
					http.MethodPatch,
					usersEndpoint+"/"+user2.ID.String(),
					models.UpdateUserRequest{},
					&server.HTTPResponse{Data: &respUserData},
				)
				s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
			})

			s.Run("404/StatusNotFound/not found", func() {
				var respUserData models.User
				resp := s.sendRequest(
					context.Background(),
					http.MethodPatch,
					usersEndpoint+"/"+uuid.New().String(),
					user3,
					&server.HTTPResponse{Data: &respUserData},
				)
				s.Require().Equal(http.StatusNotFound, resp.StatusCode)
			})
		})

		s.Run("DELETE:", func() {
			s.Run("404/not found", func() {
				var respUserData models.User
				resp := s.sendRequest(
					context.Background(),
					http.MethodDelete,
					usersEndpoint+"/"+uuid.New().String(),
					user3,
					&server.HTTPResponse{Data: &respUserData},
				)
				s.Require().Equal(http.StatusNotFound, resp.StatusCode)
			})

			s.Run("204/StatusNoContent", func() {
				resp := s.sendRequest(
					context.Background(),
					http.MethodDelete,
					usersEndpoint+"/"+user2.ID.String(),
					nil,
					nil,
				)
				s.Require().Equal(http.StatusNoContent, resp.StatusCode)
			})
		})
	})
}

func newString(value string) *string {
	return &value
}
