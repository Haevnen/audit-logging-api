package log

import (
	"context"

	entitylog "github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/interactor"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/Haevnen/audit-logging-api/internal/service"
	"github.com/Haevnen/audit-logging-api/pkg/logger"
	"github.com/google/uuid"
)

type CreateLogUseCase struct {
	Repo                repository.LogRepository
	TxManager           *interactor.TxManager
	OpenSearchPublisher service.OpenSearchPublisher
}

func NewCreateLogUseCase(repo repository.LogRepository, txManager *interactor.TxManager, openSearch service.OpenSearchPublisher) *CreateLogUseCase {
	return &CreateLogUseCase{Repo: repo, TxManager: txManager, OpenSearchPublisher: openSearch}
}

func (uc *CreateLogUseCase) Execute(ctx context.Context, log entitylog.Log) (*entitylog.Log, error) {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	if err := uc.Repo.Create(ctx, &log); err != nil {
		return nil, err
	}

	go func(l entitylog.Log) {
		if err := uc.OpenSearchPublisher.IndexLog(context.Background(), l); err != nil {
			logger.GetLogger().Errorf("failed to index log to opensearch: %v", err)
		}
	}(log)

	return &log, nil
}

func (uc *CreateLogUseCase) ExecuteBulk(ctx context.Context, logs []entitylog.Log) ([]entitylog.Log, error) {
	for i := range logs {
		if logs[i].ID == "" {
			logs[i].ID = uuid.New().String()
		}
	}

	if err := uc.TxManager.TransactionExec(ctx, func(txCtx context.Context) error {
		db := uc.TxManager.GetTx(txCtx)
		return uc.Repo.CreateBulk(txCtx, db, logs)
	}); err != nil {
		return nil, err
	}

	go func(logs []entitylog.Log) {
		if err := uc.OpenSearchPublisher.IndexLogsBulk(context.Background(), logs); err != nil {
			logger.GetLogger().Errorf("failed to index logs bulk to opensearch: %v", err)
		}
	}(logs)

	return logs, nil

}
