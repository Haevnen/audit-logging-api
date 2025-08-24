package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"

	"github.com/Haevnen/audit-logging-api/pkg/gormdb"
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

	SqsLogCleanupQueueURL  string `env:"SQS_LOG_CLEANUP_QUEUE_URL"`
	SqsLogArchivalQueueURL string `env:"SQS_LOG_ARCHIVAL_QUEUE_URL"`
	SqsIndexQueueURL       string `env:"SQS_INDEX_QUEUE_URL"`
	S3ArchiveLogURL        string `env:"S3_ARCHIVE_LOG_URL"`
	S3ArchiveLogBucketName string `env:"S3_ARCHIVE_LOG_BUCKET_NAME"`

	AwsRegion         string `env:"AWS_REGION"`
	AwsKey            string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecret         string `env:"AWS_SECRET_ACCESS_KEY"`
	LocalStackBaseURL string `env:"LOCALSTACK_BASE_URL"`

	OpenSearchURL string `env:"OPENSEARCH_URL"`
	RedisAddr     string `env:"REDIS_ADDR"`
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
