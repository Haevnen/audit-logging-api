package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSPublisher interface {
	PublishArchiveMessage(ctx context.Context, taskId string, beforeDate time.Time) error
	PublishCleanUpMessage(ctx context.Context, taskId string, beforeDate time.Time) error
	PublishIndexMessage(ctx context.Context, taskId string, logs []log.Log) error
	ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32, waitTimeSeconds int32) ([]ReceiveMessage, error)
	DeleteMessage(ctx context.Context, queueURL string, receiptHandle *string) error
}

type Message struct {
	ID         string
	BeforeDate *time.Time
	Logs       *[]log.Log
}

type ReceiveMessage struct {
	Message       Message
	ReceiveHandle *string
}

type SQSPublisherImpl struct {
	sqsClient       *sqs.Client
	archiveQueueURL string
	cleanUpQueueURL string
	indexQueueURL   string
}

func NewSQSPublisherImpl(sqsClient *sqs.Client, archiveQueueURL string, cleanUpQueueURL string, indexQueueURL string) *SQSPublisherImpl {
	return &SQSPublisherImpl{
		sqsClient:       sqsClient,
		archiveQueueURL: archiveQueueURL,
		cleanUpQueueURL: cleanUpQueueURL,
		indexQueueURL:   indexQueueURL,
	}
}

func (p *SQSPublisherImpl) PublishArchiveMessage(ctx context.Context, taskId string, beforeDate time.Time) error {
	return p.sendMessage(ctx, p.archiveQueueURL, Message{
		ID:         taskId,
		BeforeDate: &beforeDate,
	})
}

func (p *SQSPublisherImpl) PublishCleanUpMessage(ctx context.Context, taskId string, beforeDate time.Time) error {
	return p.sendMessage(ctx, p.cleanUpQueueURL, Message{
		ID:         taskId,
		BeforeDate: &beforeDate,
	})
}

func (p *SQSPublisherImpl) PublishIndexMessage(ctx context.Context, taskId string, logs []log.Log) error {
	return p.sendMessage(ctx, p.indexQueueURL, Message{
		ID:   taskId,
		Logs: &logs,
	})
}

func (p *SQSPublisherImpl) sendMessage(ctx context.Context, queueURL string, msg Message) error {
	msgBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err := p.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(msgBody)),
		QueueUrl:    aws.String(queueURL),
	}); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (p *SQSPublisherImpl) DeleteMessage(ctx context.Context, queueURL string, receiptHandle *string) error {
	if _, err := p.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: receiptHandle,
	}); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

func (p *SQSPublisherImpl) ReceiveMessages(ctx context.Context, queueURL string, maxMessages int32, waitTimeSeconds int32) ([]ReceiveMessage, error) {
	out, err := p.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     waitTimeSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	resp := make([]ReceiveMessage, 0, len(out.Messages))
	for _, msg := range out.Messages {
		var m Message
		if err := json.Unmarshal([]byte(*msg.Body), &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		resp = append(resp, ReceiveMessage{
			Message:       m,
			ReceiveHandle: msg.ReceiptHandle,
		})
	}
	return resp, nil
}
