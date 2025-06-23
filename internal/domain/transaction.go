package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	Debit  TransactionType = "DEBIT"
	Credit TransactionType = "CREDIT"
)

type SystemTransaction struct {
	ID              string
	Amount          decimal.Decimal
	Type            TransactionType
	TransactionTime time.Time
}

type BankTransaction struct {
	ID       string
	Amount   decimal.Decimal
	Date     time.Time
	BankName string
}
