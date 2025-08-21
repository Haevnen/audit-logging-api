package log

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/repository"
)

type GetLogUseCase struct {
	Repo repository.LogRepository
}

func NewGetLogUseCase(repo repository.LogRepository) *GetLogUseCase {
	return &GetLogUseCase{Repo: repo}
}

func (uc *GetLogUseCase) Execute(ctx context.Context, id string) (*log.Log, error) {
	return uc.Repo.GetByID(ctx, id)
}
