package service

import (
	"sort"
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

	for key, systemTxs := range systemTxMap {
		bankTxs, found := bankTxMap[key]
		if found {
			sort.Slice(systemTxs, func(i, j int) bool {
				return systemTxs[i].Amount.LessThan(systemTxs[j].Amount)
			})
			sort.Slice(bankTxs, func(i, j int) bool {
				return bankTxs[i].Amount.Abs().LessThan(bankTxs[j].Amount.Abs())
			})

			matchCount := min(len(systemTxs), len(bankTxs))
			summary.MatchedTransactions += matchCount

			for i := 0; i < matchCount; i++ {
				systemTx := systemTxs[i]
				bankTx := bankTxs[i]
				discrepancy := systemTx.Amount.Sub(bankTx.Amount.Abs()).Abs()
				summary.AmountDiscrepancyTotal = summary.AmountDiscrepancyTotal.Add(discrepancy)
			}

			if len(systemTxs) > matchCount {
				for _, tx := range systemTxs[matchCount:] {
					summary.UnmatchedSystemTransactions = append(summary.UnmatchedSystemTransactions, *tx)
				}
			}

			if len(bankTxs) > matchCount {
				for _, tx := range bankTxs[matchCount:] {
					summary.UnmatchedBankTransactions[tx.BankName] = append(summary.UnmatchedBankTransactions[tx.BankName], *tx)
				}
			}
			delete(bankTxMap, key)
		} else {
			for _, tx := range systemTxs {
				summary.UnmatchedSystemTransactions = append(summary.UnmatchedSystemTransactions, *tx)
			}
		}
	}

	for _, bankTxs := range bankTxMap {
		for _, tx := range bankTxs {
			summary.UnmatchedBankTransactions[tx.BankName] = append(summary.UnmatchedBankTransactions[tx.BankName], *tx)
		}
	}

	return summary
}
