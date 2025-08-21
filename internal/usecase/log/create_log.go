package log

import (
	"context"

	"github.com/Haevnen/audit-logging-api/internal/entity/log"
	"github.com/Haevnen/audit-logging-api/internal/interactor"
	"github.com/Haevnen/audit-logging-api/internal/repository"
	"github.com/google/uuid"
)

type CreateLogUseCase struct {
	Repo      repository.LogRepository
	TxManager *interactor.TxManager
}

func NewCreateLogUseCase(repo repository.LogRepository, txManager *interactor.TxManager) *CreateLogUseCase {
	return &CreateLogUseCase{Repo: repo, TxManager: txManager}
}

func (uc *CreateLogUseCase) Execute(ctx context.Context, log log.Log) (*log.Log, error) {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	if err := uc.Repo.Create(ctx, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

func (uc *CreateLogUseCase) ExecuteBulk(ctx context.Context, logs []log.Log) ([]log.Log, error) {
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

	return logs, nil

}
