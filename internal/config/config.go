package config

import (
	"fmt"

	"github.com/Haevnen/audit-logging-api/pkg/gormdb"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type Config struct {
	PostgresPort          int    `env:"POSTGRES_PORT"`
	PostgresHost          string `env:"POSTGRES_HOST"`
	PostgresUser          string `env:"POSTGRES_USER"`
	PostgresPassword      string `env:"POSTGRES_PASSWORD"`
	PostgresDB            string `env:"POSTGRES_DB"`
	PostgresContainerName string `env:"POSTGRESSQL_CONTAINER_NAME"`
	MaxOpenConn           int    `env:"POSTGRESSQL_MAX_OPEN_CONN"`
	MaxIdleConn           int    `env:"POSTGRESSQL_MAX_IDLE_CONN"`

	ProjectName string `env:"PROJECT_NAME"`
	SpecDir     string `env:"SPEC_DIR"`
	LogFile     string `env:"LOG_FILE"`

	APIPort           int    `env:"API_PORT"`
	APIHost           string `env:"API_HOST"`
	Mode              string `env:"RUN_MODE"`
	TokenSymmetricKey string `env:"TOKEN_SYMMETRIC_KEY"`

	RateLimitBurst int `env:"RATE_LIMIT_BURST"`
	RateLimitRPS   int `env:"RATE_LIMIT_RPS"`
}

func LoadConfig() (config Config, err error) {
	// Load the .env file
	err = godotenv.Load()
	if err != nil {
		panic(err)
	}

	err = env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return
}

func (e *Config) GetGORMConfig() *gormdb.Config {
	return &gormdb.Config{
		DBHost:            e.PostgresHost,
		DBPort:            e.PostgresPort,
		DBUser:            e.PostgresUser,
		DBPass:            e.PostgresPassword,
		DBName:            e.PostgresDB,
		DBLocation:        "UTC",
		LogSQL:            true,
		MaxOpenConn:       e.MaxOpenConn,
		MaxIdleConn:       e.MaxIdleConn,
		MaxLifetimeSecond: 300,
	}
}

// GetURLBase build server config from env
func (e *Config) GetURLBase() string {
	return fmt.Sprintf("%s:%d", e.APIHost, e.APIPort)
}
