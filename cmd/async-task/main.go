package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/Haevnen/audit-logging-api/internal/config"
	"github.com/Haevnen/audit-logging-api/internal/registry"
	"github.com/Haevnen/audit-logging-api/internal/worker"
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

	r := registry.NewRegistry(
		db,
		cfg.TokenSymmetricKey,
		sqsClient,
		s3Client,
		cfg.SqsLogArchivalQueueURL,
		cfg.SqsLogCleanupQueueURL,
		cfg.S3ArchiveLogBucketName,
	)

	archWorker := worker.NewArchiveWorker(
		r.QueuePublisher(),
		r.AsyncTaskRepository(),
		r.LogRepository(),
		r.S3Publisher(),
		r.TxManager(),
		cfg.SqsLogArchivalQueueURL,
	)

	cleanWorker := worker.NewCleanUpWorker(
		r.QueuePublisher(),
		r.AsyncTaskRepository(),
		r.LogRepository(),
		r.TxManager(),
		cfg.SqsLogCleanupQueueURL,
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		archWorker.Start(ctx)
	}()

	go func() {
		cleanWorker.Start(ctx)
	}()

	<-sigChan
	logger.Info("Shutting down gracefully...")
	cancel() // signal worker to stop

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
