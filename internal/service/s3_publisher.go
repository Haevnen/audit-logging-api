package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Publisher interface {
	UploadLogs(ctx context.Context, taskId string, logs []log.Log) error
}

type S3PublisherImpl struct {
	s3Client   *s3.Client
	bucketName string
}

func NewS3PublisherImpl(s3Client *s3.Client, bucketName string) *S3PublisherImpl {
	return &S3PublisherImpl{
		s3Client:   s3Client,
		bucketName: bucketName,
	}
}

func (s *S3PublisherImpl) UploadLogs(ctx context.Context, taskId string, logs []log.Log) error {
	// Marshal logs into JSON
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	// Compress with gzip
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return fmt.Errorf("failed to gzip data: %w", err)
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("failed to close gzip: %w", err)
	}

	// Generate archive key
	key := filepath.Join("archives", fmt.Sprintf("%s_%d.json.gz", taskId, time.Now().Unix()))

	// Upload to S3
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return fmt.Errorf("failed to upload logs to S3: %w", err)
	}

	return nil
}
