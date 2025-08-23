package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/pkg/logger"
)

type LogSearchFilters struct {
	TenantID  *string
	UserID    *string
	Action    *string
	Resource  *string
	Severity  *string
	StartDate *string
	EndDate   *string
	Query     *string
	Page      int
	PageSize  int
}

type SearchResult struct {
	Total int64
	Logs  []log.Log
}

type LogSearchRepository interface {
	Search(ctx context.Context, filters LogSearchFilters) (*SearchResult, error)
	Stream(ctx context.Context, filters LogSearchFilters, fn func(log.Log) error) error
}

type openSearchRepo struct {
	baseURL   string
	indexName string
	client    *http.Client
}

func NewLogSearchRepository(baseURL, indexName string) LogSearchRepository {
	return &openSearchRepo{
		baseURL:   baseURL,
		indexName: indexName,
		client:    &http.Client{},
	}
}

func (r *openSearchRepo) Search(ctx context.Context, filters LogSearchFilters) (*SearchResult, error) {
	url := fmt.Sprintf("%s/%s/_search", r.baseURL, r.indexName)

	from := (filters.Page - 1) * filters.PageSize
	if from < 0 {
		from = 0
	}

	// Build ES query
	query := buildQuery(filters, &from)
	payload, _ := json.Marshal(query)
	logger.GetLogger().Info(string(payload))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var errBody bytes.Buffer
		_, _ = errBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("opensearch error: %s - %s", resp.Status, errBody.String())
	}

	var res struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source log.Log `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	results := make([]log.Log, len(res.Hits.Hits))
	for i, h := range res.Hits.Hits {
		results[i] = h.Source
	}

	return &SearchResult{
		Total: res.Hits.Total.Value,
		Logs:  results,
	}, nil
}

func (r *openSearchRepo) Stream(ctx context.Context, filters LogSearchFilters, fn func(log.Log) error) error {
	size := 1000
	searchAfter := []interface{}{}
	logger := logger.GetLogger()

	for {
		query := buildQuery(filters, nil) // reuse your search filters
		query["size"] = size
		if len(searchAfter) > 0 {
			query["search_after"] = searchAfter
		}

		payload, _ := json.Marshal(query)
		logger.Info(string(payload))
		req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s/_search", r.baseURL, r.indexName), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := r.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var res struct {
			Hits struct {
				Hits []struct {
					Sort   []interface{} `json:"sort"`
					Source log.Log       `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return err
		}

		if len(res.Hits.Hits) == 0 {
			break
		}

		for _, h := range res.Hits.Hits {
			if err := fn(h.Source); err != nil {
				return err
			}
			searchAfter = h.Sort
		}
	}

	return nil
}

func buildQuery(filters LogSearchFilters, from *int) map[string]interface{} {
	query := map[string]interface{}{
		"size": filters.PageSize,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   []map[string]interface{}{},
				"filter": []map[string]interface{}{},
			},
		},
	}

	if from != nil {
		query["from"] = *from
	}

	query["sort"] = []map[string]interface{}{
		{"EventTimestamp": map[string]string{"order": "asc"}},
		{"ID.keyword": map[string]string{"order": "asc"}}, // tie-breaker
	}

	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})

	if filters.TenantID != nil && *filters.TenantID != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]map[string]interface{}), map[string]interface{}{
			"term": map[string]interface{}{"TenantID.keyword": *filters.TenantID},
		})
	}
	if filters.UserID != nil && *filters.UserID != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]map[string]interface{}), map[string]interface{}{
			"term": map[string]interface{}{"UserID.keyword": *filters.UserID},
		})
	}
	if filters.Action != nil && *filters.Action != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]map[string]interface{}), map[string]interface{}{
			"term": map[string]interface{}{"Action.keyword": *filters.Action},
		})
	}
	if filters.Severity != nil && *filters.Severity != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]map[string]interface{}), map[string]interface{}{
			"term": map[string]interface{}{"Severity.keyword": *filters.Severity},
		})
	}
	if filters.StartDate != nil && filters.EndDate != nil {
		gte := *filters.StartDate
		lte := *filters.EndDate

		// If the date string has no time (just YYYY-MM-DD), expand it
		if len(gte) == 10 { // "YYYY-MM-DD"
			gte += "T00:00:00Z"
		}
		if len(lte) == 10 {
			lte += "T23:59:59Z"
		}
		boolQuery["filter"] = append(boolQuery["filter"].([]map[string]interface{}), map[string]interface{}{
			"range": map[string]interface{}{
				"EventTimestamp": map[string]interface{}{
					"gte": gte,
					"lte": lte,
				},
			},
		})
	}
	if filters.Query != nil && *filters.Query != "" {
		boolQuery["must"] = append(
			boolQuery["must"].([]map[string]interface{}),
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": *filters.Query,
					"fields": []string{
						"Message", // match your mapping exactly
						"Metadata.*",
						"BeforeState.*",
						"AfterState.*",
					},
					"type":     "best_fields",
					"operator": "and",
				},
			},
		)
	}

	return query
}
