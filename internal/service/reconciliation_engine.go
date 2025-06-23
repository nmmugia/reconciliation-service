package service

import (
	"time"

	"github.com/nmmugia/reconciliation-service/internal/domain"

	"github.com/shopspring/decimal"
)

type ReconciliationEngine struct{}

func NewReconciliationEngine() *ReconciliationEngine {
	return &ReconciliationEngine{}
}

func (e *ReconciliationEngine) getMatchKey(date time.Time, txType domain.TransactionType) string {
	return date.Truncate(24*time.Hour).Format("2006-01-02") + ":" + string(txType)
}

func (e *ReconciliationEngine) Reconcile(systemTxs []domain.SystemTransaction, bankTxs []domain.BankTransaction) *domain.ReconciliationSummary {
	discrepancyThreshold := decimal.NewFromInt(1000)

	systemTxMap := make(map[string][]*domain.SystemTransaction)
	for i := range systemTxs {
		tx := &systemTxs[i]
		key := e.getMatchKey(tx.TransactionTime, tx.Type)
		systemTxMap[key] = append(systemTxMap[key], tx)
	}

	bankTxMap := make(map[string][]*domain.BankTransaction)
	for i := range bankTxs {
		tx := &bankTxs[i]
		txType := domain.Credit
		if tx.Amount.IsNegative() {
			txType = domain.Debit
		}
		key := e.getMatchKey(tx.Date, txType)
		bankTxMap[key] = append(bankTxMap[key], tx)
	}

	summary := &domain.ReconciliationSummary{
		UnmatchedBankTransactions:   make(map[string][]domain.BankTransaction),
		UnmatchedSystemTransactions: make([]domain.SystemTransaction, 0),
		AmountDiscrepancyTotal:      decimal.Zero,
	}
	summary.TotalSystemTransactions = len(systemTxs)
	summary.TotalBankTransactions = len(bankTxs)

	processedSystemTx := make(map[string]bool)
	processedBankTx := make(map[string]bool)

	for key, systemTxs := range systemTxMap {
		bankTxs, found := bankTxMap[key]
		if !found {
			continue
		}

		bankTxUsed := make([]bool, len(bankTxs))

		for _, systemTx := range systemTxs {
			bestFitIndex := -1
			minDifference := discrepancyThreshold

			for i, bankTx := range bankTxs {
				if bankTxUsed[i] {
					continue
				}

				currentDifference := systemTx.Amount.Sub(bankTx.Amount.Abs()).Abs()

				if currentDifference.LessThan(minDifference) {
					minDifference = currentDifference
					bestFitIndex = i
				}
			}

			if bestFitIndex != -1 {
				summary.MatchedTransactions++
				summary.AmountDiscrepancyTotal = summary.AmountDiscrepancyTotal.Add(minDifference)

				processedSystemTx[systemTx.ID] = true
				processedBankTx[bankTxs[bestFitIndex].ID] = true
				bankTxUsed[bestFitIndex] = true
			}
		}
	}

	for _, tx := range systemTxs {
		if !processedSystemTx[tx.ID] {
			summary.UnmatchedSystemTransactions = append(summary.UnmatchedSystemTransactions, tx)
		}
	}
	for _, tx := range bankTxs {
		if !processedBankTx[tx.ID] {
			summary.UnmatchedBankTransactions[tx.BankName] = append(summary.UnmatchedBankTransactions[tx.BankName], tx)
		}
	}

	return summary
}
