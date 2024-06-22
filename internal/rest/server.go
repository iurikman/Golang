package server

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

const (
	readHeaderTimeout       = 5
	gracefulShutdownTimeout = 10 * time.Second
)

type Server struct {
	router  *chi.Mux
	cfg     Config
	key     *rsa.PublicKey
	service service
	server  *http.Server
}

type Config struct {
	BindAddress string
}

func NewServer(cfg Config, service service, key *rsa.PublicKey) *Server {
	router := chi.NewRouter()

	return &Server{
		cfg:     cfg,
		router:  router,
		service: service,
		server: &http.Server{
			Addr:              cfg.BindAddress,
			ReadHeaderTimeout: readHeaderTimeout * time.Second,
			Handler:           router,
		},
		key: key,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.configRouter()

	go func() {
		<-ctx.Done()
		ctxWithTimeOut, cancel := context.WithTimeout(ctx, gracefulShutdownTimeout)

		defer cancel()

		err := s.server.Shutdown(ctxWithTimeOut)
		if err != nil {
			logrus.Warnf("server Shutdown error: %v", err)
		}
	}()

	err := s.server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server closed error: %w", err)
	}

	return nil
}

func (s *Server) configRouter() {
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(s.JWTAuth)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", s.createUser)
			r.Get("/{id}", s.getUserByID)
			r.Get("/", s.getUsers)
			r.Patch("/{id}", s.updateUser)
			r.Delete("/{id}", s.deleteUser)
		})

		r.Route("/companies", func(r chi.Router) {
			r.Post("/", s.createCompany)
			r.Get("/", s.getCompanies)
			r.Patch("/{id}", s.updateCompany)
		})

		r.Route("/storage", func(r chi.Router) {
			r.Post("/", s.uploadFile)
			r.Get("/{id}", s.getFile)
		})
	})
}
