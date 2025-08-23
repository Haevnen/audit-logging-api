package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/redis/go-redis/v9"
)

type PubSub interface {
	Publish(ctx context.Context, channel string, message string) error
	Subscribe(ctx context.Context, channel string) *redis.PubSub
	BroadcastLogs(ctx context.Context, logs []log.Log) error
	BroadcastLog(ctx context.Context, logRecord log.Log) error
}

type PubSubImpl struct {
	client *redis.Client
}

func NewPubSubImpl(addr string) *PubSubImpl {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, // e.g. "redis:6379"
	})
	return &PubSubImpl{client: rdb}
}

func (r *PubSubImpl) Publish(ctx context.Context, channel string, message string) error {
	return r.client.Publish(ctx, channel, message).Err()
}

func (r *PubSubImpl) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.client.Subscribe(ctx, channel)
}

// BroadcastLogs publishes multiple logs at once
func (r *PubSubImpl) BroadcastLogs(ctx context.Context, logs []log.Log) error {
	for _, l := range logs {
		if err := r.BroadcastLog(ctx, l); err != nil {
			return err
		}
	}
	return nil
}

// BroadcastLog marshals and publishes a log record
func (r *PubSubImpl) BroadcastLog(ctx context.Context, logRecord log.Log) error {
	payload, err := json.Marshal(logRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	// Tenant-specific channel
	if len(logRecord.TenantID) > 0 {
		tenantChannel := "logs:" + logRecord.TenantID
		if err := r.Publish(ctx, tenantChannel, string(payload)); err != nil {
			return fmt.Errorf("publish to tenant channel: %w", err)
		}
	}

	// Global channel (optional, e.g. admin subscribers)
	if err := r.Publish(ctx, "logs", string(payload)); err != nil {
		return fmt.Errorf("publish to global channel: %w", err)
	}

	return nil
}
