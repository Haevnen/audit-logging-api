package service

//go:generate mockgen -source=opensearch_publisher.go -destination=./mocks/mock_opensearch_publisher.go -package=mocks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
)

type OpenSearchPublisher interface {
	IndexLog(ctx context.Context, l log.Log) error
	IndexLogsBulk(ctx context.Context, logs []log.Log) error
	DeleteLogsBulk(ctx context.Context, ids []string) error
}

type openSearchPublisher struct {
	baseURL   string
	indexName string
	client    *http.Client
}

func NewOpenSearchPublisher(baseURL, indexName string) OpenSearchPublisher {
	return &openSearchPublisher{
		baseURL:   baseURL,
		indexName: indexName,
		client:    &http.Client{},
	}
}

func (p *openSearchPublisher) IndexLog(ctx context.Context, l log.Log) error {
	url := fmt.Sprintf("%s/%s/_doc/%s", p.baseURL, p.indexName, l.ID)

	body, err := json.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshal log: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("opensearch index error: status %s", resp.Status)
	}
	return nil
}

func (p *openSearchPublisher) IndexLogsBulk(ctx context.Context, logs []log.Log) error {
	url := fmt.Sprintf("%s/%s/_bulk", p.baseURL, p.indexName)

	var buf bytes.Buffer
	for _, l := range logs {
		meta := fmt.Sprintf(`{ "index": { "_id": "%s" } }%s`, l.ID, "\n")
		body, err := json.Marshal(l)
		if err != nil {
			return fmt.Errorf("marshal bulk log: %w", err)
		}
		buf.WriteString(meta)
		buf.Write(body)
		buf.WriteString("\n")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return fmt.Errorf("new bulk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send bulk request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("bulk opensearch index error: status %s", resp.Status)
	}
	return nil
}

// DeleteLogsBulk deletes multiple documents from OpenSearch by ID in chunks
func (p *openSearchPublisher) DeleteLogsBulk(ctx context.Context, ids []string) error {
	const batchSize = 1000

	for start := 0; start < len(ids); start += batchSize {
		end := start + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[start:end]

		if err := p.deleteLogsChunk(ctx, batch); err != nil {
			return fmt.Errorf("delete logs chunk [%d:%d]: %w", start, end, err)
		}
	}

	return nil
}

func (p *openSearchPublisher) deleteLogsChunk(ctx context.Context, ids []string) error {
	url := fmt.Sprintf("%s/%s/_bulk", p.baseURL, p.indexName)

	var buf bytes.Buffer
	for _, id := range ids {
		meta := fmt.Sprintf(`{ "delete": { "_id": "%s" } }%s`, id, "\n")
		buf.WriteString(meta)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("new bulk delete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("send bulk delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("bulk opensearch delete error: status %s", resp.Status)
	}
	return nil
}
