package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"

	handler "github.com/Haevnen/audit-logging-api/internal/adapter/http"
	api_service "github.com/Haevnen/audit-logging-api/internal/adapter/http/gen/api"
	"github.com/Haevnen/audit-logging-api/internal/config"
	"github.com/Haevnen/audit-logging-api/internal/infra/middleware"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/pkg/gormdb"
	"github.com/Haevnen/audit-logging-api/pkg/logger"
)

func start() int {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	logger := logger.InitLogger(cfg.Mode, cfg.LogFile)

	// Connect to database
	gormCfg := cfg.GetGORMConfig()
	db, closeFunc, err := gormdb.New(gormCfg)
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to connect to database")
		return 1
	}
	defer func() {
		if err := closeFunc(); err != nil {
			logger.WithField("error", err).Fatal("Failed to close connection to database")
		}
	}()

	// Setup router
	r := gin.Default()
	gin.SetMode(cfg.Mode)

	sqsClient, err := NewSQSClient(cfg)
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to create SQS client")
		return 1
	}

	s3Client, err := NewS3Client(cfg)
	if err != nil {
		logger.WithField("error", err).Fatal("Failed to create S3 client")
		return 1
	}

	registry := registry.NewRegistry(
		db,
		cfg.TokenSymmetricKey,
		sqsClient,
		s3Client,
		cfg.SqsLogArchivalQueueURL,
		cfg.SqsLogCleanupQueueURL,
		cfg.SqsIndexQueueURL,
		cfg.S3ArchiveLogBucketName,
		cfg.OpenSearchURL,
		cfg.RedisAddr,
	)
	handler := handler.New(registry)
	jwt := registry.Manager()

	// Register health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Register handlers
	api_service.RegisterHandlersWithOptions(r, handler, api_service.GinServerOptions{
		BaseURL: "/api/v1",
		Middlewares: []api_service.MiddlewareFunc{
			middleware.RequireAuth(jwt),
			middleware.RequireRole(),
			middleware.RequireRateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst),
		},
	})

	// Start server
	s := &http.Server{
		Addr:    cfg.GetURLBase(),
		Handler: r,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithField("error", err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown:", err)
		return 1
	}

	select {
	case <-ctx.Done():
		logger.Info("timeout of 5 seconds")
	}
	logger.Info("Server Exited")
	return 0
}

func NewSQSClient(env config.Config) (*sqs.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion(env.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			env.AwsKey, env.AwsSecret, "",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	sqsClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(env.LocalStackBaseURL)
	})
	return sqsClient, nil
}

func NewS3Client(env config.Config) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion(env.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			env.AwsKey, env.AwsSecret, "",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(env.LocalStackBaseURL) // Localstack support
		o.UsePathStyle = true
	})
	return s3Client, nil
}

func main() {
	start()
}
