package pgstore

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
)

type Postgres struct {
	db  *pgxpool.Pool
	dsn string
}

//go:embed migrations
var migrations embed.FS

type Config struct {
	PGUser     string
	PGPassword string
	PGHost     string
	PGPort     string
	PGDatabase string
}

func New(ctx context.Context, cfg Config) (*Postgres, error) {
	urlScheme := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.PGUser, cfg.PGPassword),
		Host:     fmt.Sprintf("%s:%s", cfg.PGHost, cfg.PGPort),
		Path:     cfg.PGDatabase,
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	dsn := urlScheme.String()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New(ctx, dsn): %w", err)
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	log.Info("successfully connected to db")

	return &Postgres{
		db:  db,
		dsn: dsn,
	}, nil
}

func (p *Postgres) Migrate(direction migrate.MigrationDirection) error {
	conn, err := sql.Open("pgx", p.dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warnf("conn.Close: %v", err)
		}
	}()

	assetDir := func() func(string) ([]string, error) {
		return func(path string) ([]string, error) {
			dirEntry, err := migrations.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("migrations.ReadDir: %w", err)
			}

			entries := make([]string, 0)

			for _, e := range dirEntry {
				entries = append(entries, e.Name())
			}

			return entries, nil
		}
	}()

	asset := migrate.AssetMigrationSource{
		Asset:    migrations.ReadFile,
		AssetDir: assetDir,
		Dir:      "migrations",
	}

	_, err = migrate.Exec(conn, "postgres", asset, direction)
	if err != nil {
		return fmt.Errorf("migrate.Exec: %w", err)
	}

	return nil
}

func (p *Postgres) Truncate(ctx context.Context, tables ...string) error {
	for _, table := range tables {
		_, err := p.db.Exec(ctx, "DELETE FROM "+table)
		if err != nil {
			return fmt.Errorf("truncate: %w", err)
		}
	}

	return nil
}
