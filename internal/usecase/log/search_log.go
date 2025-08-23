package log

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/repository"
)

type SearchLogsUseCase struct {
	Repo repository.LogSearchRepository
}

func NewSearchLogsUseCase(repo repository.LogSearchRepository) *SearchLogsUseCase {
	return &SearchLogsUseCase{Repo: repo}
}

func (uc *SearchLogsUseCase) Execute(ctx context.Context, filters repository.LogSearchFilters) (*repository.SearchResult, error) {
	return uc.Repo.Search(ctx, filters)
}

func (uc *SearchLogsUseCase) Stream(ctx context.Context, filters repository.LogSearchFilters, fn func(log.Log) error) error {
	return uc.Repo.Stream(ctx, filters, fn)
}
