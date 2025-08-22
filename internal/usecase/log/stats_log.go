package log

import (
	"context"
	"time"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/repository"
)

type GetStatsUseCase struct {
	Repo repository.LogRepository
}

func NewGetStatsUseCase(repo repository.LogRepository) *GetStatsUseCase {
	return &GetStatsUseCase{Repo: repo}
}

func (uc *GetStatsUseCase) Execute(ctx context.Context, tenantId string, startTime, endTime time.Time) ([]log.LogStats, error) {
	return uc.Repo.GetStats((ctx), tenantId, startTime, endTime)
}
