package usecase

import (
	"context"
	"time"

	"github.com/nmmugia/reconciliation-service/internal/domain"
)

type DataRepository interface {
	LoadSystemTransactions(ctx context.Context, filePath string, startDate, endDate time.Time) ([]domain.SystemTransaction, error)
	LoadBankStatements(ctx context.Context, filePaths []string, startDate, endDate time.Time) ([]domain.BankTransaction, error)
}

type Reconciler interface {
	Reconcile(ctx context.Context, systemFile string, bankFiles []string, startDate, endDate time.Time) (*domain.ReconciliationSummary, error)
}
