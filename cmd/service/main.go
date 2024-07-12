package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/iurikman/smartSurvey/internal/config"
	minio2 "github.com/iurikman/smartSurvey/internal/filestore"
	"github.com/iurikman/smartSurvey/internal/jwtgenerator"
	"github.com/iurikman/smartSurvey/internal/logger"
	"github.com/iurikman/smartSurvey/internal/pgstore"
	server "github.com/iurikman/smartSurvey/internal/rest"
	"github.com/iurikman/smartSurvey/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.InitLogger("debug")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	defer cancel()

	cfg := config.New()

	pgStore, err := pgstore.New(ctx, pgstore.Config{
		PGUser:     cfg.PGUser,
		PGPassword: cfg.PGPassword,
		PGHost:     cfg.PGHost,
		PGPort:     cfg.PGPort,
		PGDatabase: cfg.PGDatabase,
	})
	if err != nil {
		log.Panicf("pgStore.New: %v", err)
	}

	if err := pgStore.Migrate(migrate.Up); err != nil {
		log.Panicf("pgStore.Migrate: %v", err)
	}

	log.Info("successful migration")

	// Initialize minio client object.
	minioStorage, err := minio2.NewMinioStorage("localhost:9000", "minio", "qwerqwer")
	if err != nil {
		log.Panicf("minioStorage init error: %v", err)
	}

	jwtGenerator := jwtgenerator.NewJWTGenerator()

	svc := service.New(pgStore, minioStorage)
	srv := server.NewServer(
		server.Config{BindAddress: cfg.BindAddress},
		svc,
		jwtGenerator.GetPublicKey(),
	)

	if err := srv.Start(ctx); err != nil {
		log.Panicf("Server start error: %v", err)
	}
}
