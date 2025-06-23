package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type ReconciliationSummary struct {
	TotalSystemTransactions     int
	TotalBankTransactions       int
	MatchedTransactions         int
	UnmatchedSystemTransactions []SystemTransaction
	UnmatchedBankTransactions   map[string][]BankTransaction
	AmountDiscrepancyTotal      decimal.Decimal
	ProcessingDurationSeconds   float64
}

type TransactionDataReader interface {
	ReadSystemTransactions(filePath string, startDate, endDate time.Time) ([]SystemTransaction, error)
}

type BankStatementReader interface {
	ReadBankTransactions(filePaths []string, startDate, endDate time.Time) ([]BankTransaction, error)
}
