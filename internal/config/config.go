package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	BindAddress string `env:"BIND_ADDR" env-default:":8080"`
	LogLevel    string `env:"LOG_LEVEL" env-default:"debug"`

	PGHost     string `env:"PG_HOST" env-default:"localhost"`
	PGPort     string `env:"PG_PORT" env-default:"5432"`
	PGDatabase string `env:"PG_DATABASE" env-default:"postgresdb"`
	PGUser     string `env:"PG_USER" env-default:"postgres"`
	PGPassword string `env:"PG_PASSWORD" env-default:"qwerqwer"`

	StorageHost            string `env:"STORAGE_HOST" env-default:"localhost"`
	StoragePort            string `env:"STORAGE_PORT" env-default:"9000"`
	StorageAccessKey       string `env:"STORAGE_ACCESS_KEY" env-default:"minio"`
	StorageSecretAccessKey string `env:"STORAGE_SECRET_ACCESS_KEY" env-default:"qwerqwer"`
}

func New() Config {
	cfg := Config{}

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(fmt.Sprintf("error getting config: %v", err))
	}

	return cfg
}
