package usecase

import (
	"fmt"
	"sync"
	"time"

	"github.com/nmmugia/reconciliation-service/internal/domain"
	"github.com/nmmugia/reconciliation-service/internal/service"
)

type ReconciliationUsecase struct {
	sysTxReader  domain.TransactionDataReader
	bankTxReader domain.BankStatementReader
	engine       *service.ReconciliationEngine
}

func NewReconciliationUsecase(sysReader domain.TransactionDataReader, bankReader domain.BankStatementReader, engine *service.ReconciliationEngine) *ReconciliationUsecase {
	return &ReconciliationUsecase{
		sysTxReader:  sysReader,
		bankTxReader: bankReader,
		engine:       engine,
	}
}

func (uc *ReconciliationUsecase) PerformReconciliation(sysTxPath string, bankTxPaths []string, start, end time.Time) (*domain.ReconciliationSummary, error) {
	var sysTxs []domain.SystemTransaction
	var bankTxs []domain.BankTransaction
	var sysErr, bankErr error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		sysTxs, sysErr = uc.sysTxReader.ReadSystemTransactions(sysTxPath, start, end)
	}()

	go func() {
		defer wg.Done()
		bankTxs, bankErr = uc.bankTxReader.ReadBankTransactions(bankTxPaths, start, end)
	}()

	wg.Wait()

	if sysErr != nil {
		return nil, fmt.Errorf("failed to read system transactions: %w", sysErr)
	}
	if bankErr != nil {
		return nil, fmt.Errorf("failed to read bank statements: %w", bankErr)
	}

	summary := uc.engine.Reconcile(sysTxs, bankTxs)

	return summary, nil
}
