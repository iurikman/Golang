package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	minio2 "github.com/iurikman/smartSurvey/internal/filestore"

	"github.com/iurikman/smartSurvey/internal/jwtgenerator"
	"github.com/iurikman/smartSurvey/internal/models"

	"github.com/google/uuid"

	"github.com/iurikman/smartSurvey/internal/config"
	"github.com/iurikman/smartSurvey/internal/pgstore"
	server "github.com/iurikman/smartSurvey/internal/rest"
	"github.com/iurikman/smartSurvey/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
)

const (
	bindAddr          = "http://localhost:8080/api/v1"
	usersEndpoint     = "/users"
	companiesEndpoint = "/companies"
	storageEndpoint   = "/storage"
)

type IntegrationTestSuite struct {
	suite.Suite
	cancel         context.CancelFunc
	pgStore        *pgstore.Postgres
	service        *service.Service
	server         *server.Server
	authToken      string
	testCompanyID  uuid.UUID
	testUserID     uuid.UUID
	tokenGenerator *jwtgenerator.JWTGenerator
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.testUserID = uuid.New()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	cfg := config.New()

	pgStore, err := pgstore.New(ctx, pgstore.Config{
		PGUser:     cfg.PGUser,
		PGPassword: cfg.PGPassword,
		PGHost:     cfg.PGHost,
		PGPort:     cfg.PGPort,
		PGDatabase: cfg.PGDatabase,
	})
	s.Require().NoError(err)

	s.pgStore = pgStore

	err = pgStore.Migrate(migrate.Up)
	s.Require().NoError(err)

	s.tokenGenerator = jwtgenerator.NewJWTGenerator()
	company, err := s.pgStore.CreateCompany(ctx, models.Company{
		ID:   uuid.New(),
		Name: "test company",
	})
	s.Require().NoError(err)
	user, err := s.pgStore.CreateUser(ctx, models.User{ID: s.testUserID, Email: newString("testemail@qwerty.org"), UserType: newString("testType"), Company: company.ID})
	s.Require().NoError(err)
	s.testUserID = user.ID
	s.authToken, err = s.tokenGenerator.GetNewTokenString(*user)
	s.Require().NoError(err)
	minioStorage, err := minio2.NewMinioStorage("localhost:9000", "minio", "qwerqwer")
	s.Require().NoError(err)

	s.service = service.New(pgStore, minioStorage)

	s.server = server.NewServer(
		server.Config{BindAddress: cfg.BindAddress},
		s.service,
		s.tokenGenerator.GetPublicKey(),
	)

	err = pgStore.Truncate(ctx, "users", "companies")
	s.Require().NoError(err)

	go func() {
		err = s.server.Start(ctx)
		s.Require().NoError(err)
	}()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.cancel()
}

func (s *IntegrationTestSuite) sendRequest(ctx context.Context, method, endpoint string, body interface{}, dest interface{}) *http.Response {
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
