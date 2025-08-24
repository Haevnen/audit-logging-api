package log

//go:generate mockgen -source=interface.go -destination=./mocks/mock_usecase.go -package=mocks

import (
	"context"
	"time"

	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/repository"
)

type CreateLogUseCaseInterface interface {
	Execute(ctx context.Context, tenantId, userId string, log entitylog.Log) (*entitylog.Log, error)
	ExecuteBulk(ctx context.Context, tenantId, userId string, logs []entitylog.Log) ([]entitylog.Log, error)
}

type GetLogUseCaseInterface interface {
	Execute(ctx context.Context, id, tenantId string) (*entitylog.Log, error)
}

type DeleteLogUseCaseInterface interface {
	Execute(ctx context.Context, tenantId, userId string, beforeDate time.Time) error
}

type GetStatsUseCaseInterface interface {
	Execute(ctx context.Context, tenantId string, start, end time.Time) ([]entitylog.LogStats, error)
}

type SearchLogsUseCaseInterface interface {
	Execute(ctx context.Context, filters repository.LogSearchFilters) (*repository.SearchResult, error)
	Stream(ctx context.Context, filters repository.LogSearchFilters, fn func(entitylog.Log) error) error
}
