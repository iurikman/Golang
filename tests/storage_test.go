package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/iurikman/smartSurvey/internal/models"
	server "github.com/iurikman/smartSurvey/internal/rest"
)

var (
	testID    = uuid.New()
	testName  = "userstest"
	testBytes = []byte("Hello World")
)

func (s *IntegrationTestSuite) TestStorage() {
	s.Run("POST:", func() {
		s.Run("201/statusCreated", func() {
			for i := 0; i < 12; i++ {
				respData := new(models.File)

				testFile, err := os.Open("users_test.go")
				s.Require().NoError(err)
				defer func(testFile *os.File) {
					err := testFile.Close()
					if err != nil {
						log.Warn("test file close error:", err)
					}
				}(testFile)

				bytesTestFile, err := io.ReadAll(testFile)

				resp := s.sendRequestToStorage(
					context.Background(),
					http.MethodPost,
					storageEndpoint,
					models.File{
						ID:    testID,
						Name:  testName,
						Size:  int64(len(bytesTestFile)),
						Bytes: bytesTestFile,
					},
					&server.HTTPResponse{Data: &respData},
				)
				s.Require().Equal(http.StatusCreated, resp.StatusCode)
				s.Require().NotZero(respData.ID)
				s.Require().NotNil(respData.ID)
				s.Require().Equal(testName, respData.Name)
				testID = respData.ID
				testName = respData.Name
				testBytes = bytesTestFile
			}
		})
	})

	s.Run("GET:", func() {
		s.Run("200/statusOk", func() {
			respFile := new(models.File)
			resp := s.sendRequestToStorage(
				context.Background(),
				http.MethodGet,
				storageEndpoint+"/"+testID.String()+"?bucketName="+testName,
				nil,
				&server.HTTPResponse{Data: &respFile},
			)

			s.Require().Equal(http.StatusOK, resp.StatusCode)
			s.Require().Equal(testID, respFile.ID)
			s.Require().Equal(testName, respFile.Name)
			s.Require().Equal(testBytes, respFile.Bytes)
		})

		s.Run("200/getBucketFiles", func() {
			respFiles := new([]models.File)
			resp := s.sendRequestToStorage(
				context.Background(),
				http.MethodGet,
				storageEndpoint+"/getBucketFiles/"+testName,
				nil,
				&server.HTTPResponse{Data: &respFiles},
			)
			s.Require().Equal(http.StatusOK, resp.StatusCode)
		})
	})

	s.Run("Delete:", func() {
		s.Run("200/statusOk", func() {
			resp := s.sendRequestToStorage(
				context.Background(),
				http.MethodDelete,
				storageEndpoint+"/"+testID.String()+"?bucketName="+testName,
				nil,
				nil,
			)

			s.Require().Equal(http.StatusNoContent, resp.StatusCode)
		})
	})
}

func (s *IntegrationTestSuite) sendRequestToStorage(ctx context.Context, method, endpoint string, body interface{}, dest interface{}) *http.Response {
	s.T().Helper()

	reqBody, err := json.Marshal(body)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(ctx, method, bindAddr+endpoint, bytes.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+s.authToken)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()

	if dest != nil {
		err = json.NewDecoder(resp.Body).Decode(&dest)
		s.Require().NoError(err)
	}

	return resp
}
