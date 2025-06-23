package service

import (
	"testing"
	"time"

	"github.com/nmmugia/reconciliation-service/internal/domain"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newDecimalFromString(val string) decimal.Decimal {
	d, err := decimal.NewFromString(val)
	if err != nil {
		panic(err)
	}
	return d
}

func newDate(day int) time.Time {
	return time.Date(2023, 1, day, 0, 0, 0, 0, time.UTC)
}

func TestReconciliationEngine_Reconcile(t *testing.T) {
	engine := NewReconciliationEngine()

	testCases := []struct {
		name                         string
		systemTxs                    []domain.SystemTransaction
		bankTxs                      []domain.BankTransaction
		expectedMatchedCount         int
		expectedUnmatchedSystemCount int
		expectedUnmatchedBankCount   int
		expectedDiscrepancy          decimal.Decimal
	}{
		{
			name:                         "Empty Inputs",
			systemTxs:                    []domain.SystemTransaction{},
			bankTxs:                      []domain.BankTransaction{},
			expectedMatchedCount:         0,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   0,
			expectedDiscrepancy:          newDecimalFromString("0"),
		},
		{
			name: "Perfect One-to-One Match",
			systemTxs: []domain.SystemTransaction{
				{ID: "S1", Amount: newDecimalFromString("100"), Type: domain.Credit, TransactionTime: newDate(1)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B1", Amount: newDecimalFromString("100"), Date: newDate(1), BankName: "BankA"},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   0,
			expectedDiscrepancy:          newDecimalFromString("0"),
		},
		{
			name: "No Matches - All Unmatched",
			systemTxs: []domain.SystemTransaction{
				{ID: "S1", Amount: newDecimalFromString("100"), Type: domain.Credit, TransactionTime: newDate(1)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B1", Amount: newDecimalFromString("200"), Date: newDate(2), BankName: "BankA"},
			},
			expectedMatchedCount:         0,
			expectedUnmatchedSystemCount: 1,
			expectedUnmatchedBankCount:   1,
			expectedDiscrepancy:          newDecimalFromString("0"),
		},
		{
			name: "Many-to-One Match (System Surplus)",
			systemTxs: []domain.SystemTransaction{
				{ID: "S1", Amount: newDecimalFromString("100"), Type: domain.Credit, TransactionTime: newDate(1)},
				{ID: "S2", Amount: newDecimalFromString("100.10"), Type: domain.Credit, TransactionTime: newDate(1)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B1", Amount: newDecimalFromString("100"), Date: newDate(1), BankName: "BankA"},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 1,
			expectedUnmatchedBankCount:   0,
			expectedDiscrepancy:          newDecimalFromString("0"),
		},
		{
			name: "One-to-Many Match (Bank Surplus)",
			systemTxs: []domain.SystemTransaction{
				{ID: "S1", Amount: newDecimalFromString("99.90"), Type: domain.Credit, TransactionTime: newDate(1)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B1", Amount: newDecimalFromString("100"), Date: newDate(1), BankName: "BankA"},
				{ID: "B2", Amount: newDecimalFromString("100"), Date: newDate(1), BankName: "BankA"},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   1,
			expectedDiscrepancy:          newDecimalFromString("0.10"),
		},
		{
			name: "Handles Negative Bank Amounts (DEBIT)",
			systemTxs: []domain.SystemTransaction{
				{ID: "S1", Amount: newDecimalFromString("150.25"), Type: domain.Debit, TransactionTime: newDate(1)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B1", Amount: newDecimalFromString("-150"), Date: newDate(1), BankName: "BankA"},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   0,
			expectedDiscrepancy:          newDecimalFromString("0.25"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summary := engine.Reconcile(tc.systemTxs, tc.bankTxs)

			assert.Equal(t, tc.expectedMatchedCount, summary.MatchedTransactions, "Mismatched count of matched transactions")
			assert.Equal(t, tc.expectedUnmatchedSystemCount, len(summary.UnmatchedSystemTransactions), "Mismatched count of unmatched system transactions")

			unmatchedBankCount := 0
			for _, txs := range summary.UnmatchedBankTransactions {
				unmatchedBankCount += len(txs)
			}
			assert.Equal(t, tc.expectedUnmatchedBankCount, unmatchedBankCount, "Mismatched count of unmatched bank transactions")

			assert.True(t, tc.expectedDiscrepancy.Equal(summary.AmountDiscrepancyTotal), "Expected discrepancy of %s but got %s", tc.expectedDiscrepancy.String(), summary.AmountDiscrepancyTotal.String())
		})
	}
}
