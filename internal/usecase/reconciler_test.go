package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/nmmugia/reconciliation-service/internal/domain"
	"github.com/nmmugia/reconciliation-service/internal/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockSuccessReader struct{}

func (m *mockSuccessReader) ReadSystemTransactions(filePath string, startDate, endDate time.Time) ([]domain.SystemTransaction, error) {
	return []domain.SystemTransaction{
		{ID: "sys1", Amount: decimal.NewFromInt(100), Type: domain.Credit, TransactionTime: time.Now()},
	}, nil
}
func (m *mockSuccessReader) ReadBankTransactions(filePaths []string, startDate, endDate time.Time) ([]domain.BankTransaction, error) {
	return []domain.BankTransaction{
		{ID: "bank1", Amount: decimal.NewFromInt(100), Date: time.Now()},
	}, nil
}

type mockErrorReader struct{}

func (m *mockErrorReader) ReadSystemTransactions(filePath string, startDate, endDate time.Time) ([]domain.SystemTransaction, error) {
	return nil, errors.New("mock system error")
}
func (m *mockErrorReader) ReadBankTransactions(filePaths []string, startDate, endDate time.Time) ([]domain.BankTransaction, error) {
	return nil, errors.New("mock bank error")
}

func TestReconciliationUsecase_PerformReconciliation(t *testing.T) {
	engine := service.NewReconciliationEngine()

	t.Run("successful reconciliation", func(t *testing.T) {
		uc := NewReconciliationUsecase(&mockSuccessReader{}, &mockSuccessReader{}, engine)
		summary, err := uc.PerformReconciliation("sample/system.csv", []string{"sample/bank.csv"}, time.Now(), time.Now())

		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 1, summary.MatchedTransactions)
	})

	t.Run("system reader error", func(t *testing.T) {
		uc := NewReconciliationUsecase(&mockErrorReader{}, &mockSuccessReader{}, engine)
		_, err := uc.PerformReconciliation("sample/system.csv", []string{"sample/bank.csv"}, time.Now(), time.Now())
		assert.Error(t, err)
		assert.Equal(t, "failed to read system transactions: mock system error", err.Error())
	})

	t.Run("bank reader error", func(t *testing.T) {
		uc := NewReconciliationUsecase(&mockSuccessReader{}, &mockErrorReader{}, engine)
		_, err := uc.PerformReconciliation("sample/system.csv", []string{"sample/bank.csv"}, time.Now(), time.Now())
		assert.Error(t, err)
		assert.Equal(t, "failed to read bank statements: mock bank error", err.Error())
	})
}
