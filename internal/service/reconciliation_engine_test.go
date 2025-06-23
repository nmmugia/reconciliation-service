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

func TestReconciliationEngine_Reconcile_BestFit(t *testing.T) {
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
			name: "Perfect Match",
			systemTxs: []domain.SystemTransaction{
				{ID: "S1", Amount: newDecimalFromString("100"), Type: domain.Credit, TransactionTime: newDate(1)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B1", Amount: newDecimalFromString("100"), Date: newDate(1)},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   0,
			expectedDiscrepancy:          newDecimalFromString("0"),
		},
		{
			name: "Simple Discrepancy",
			systemTxs: []domain.SystemTransaction{
				{ID: "S2", Amount: newDecimalFromString("100.50"), Type: domain.Debit, TransactionTime: newDate(2)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B2", Amount: newDecimalFromString("-100.00"), Date: newDate(2)},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   0,
			expectedDiscrepancy:          newDecimalFromString("0.50"),
		},
		{
			name: "Best Fit pairing logic test",
			systemTxs: []domain.SystemTransaction{
				{ID: "S3", Amount: newDecimalFromString("75500.50"), Type: domain.Debit, TransactionTime: newDate(3)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B3-FEE", Amount: newDecimalFromString("-15000.00"), Date: newDate(3)},
				{ID: "B3-REAL", Amount: newDecimalFromString("-75500.00"), Date: newDate(3)},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   1,
			expectedDiscrepancy:          newDecimalFromString("0.50"),
		},
		{
			name: "Threshold Test - difference is too large to match",
			systemTxs: []domain.SystemTransaction{
				{ID: "S4", Amount: newDecimalFromString("10000"), Type: domain.Debit, TransactionTime: newDate(4)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B4", Amount: newDecimalFromString("-8000"), Date: newDate(4)},
			},
			expectedMatchedCount:         0,
			expectedUnmatchedSystemCount: 1,
			expectedUnmatchedBankCount:   1,
			expectedDiscrepancy:          newDecimalFromString("0"),
		},
		{
			name: "Full real-world scenario from test failure",
			systemTxs: []domain.SystemTransaction{
				{ID: "SYS-001", Amount: newDecimalFromString("50000.00"), Type: domain.Credit, TransactionTime: newDate(23)},
				{ID: "SYS-002", Amount: newDecimalFromString("125000.00"), Type: domain.Debit, TransactionTime: newDate(23)},
				{ID: "SYS-003", Amount: newDecimalFromString("75500.50"), Type: domain.Debit, TransactionTime: newDate(24)},
				{ID: "SYS-004", Amount: newDecimalFromString("200000.00"), Type: domain.Credit, TransactionTime: newDate(24)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "BNK-A-100", Amount: newDecimalFromString("50000.00"), Date: newDate(23)},
				{ID: "BNK-A-101", Amount: newDecimalFromString("-125500.00"), Date: newDate(23)},
				{ID: "BNK-A-102", Amount: newDecimalFromString("-75500.50"), Date: newDate(24)},
				{ID: "BNK-FEE-X", Amount: newDecimalFromString("-15000.00"), Date: newDate(24)},
			},
			expectedMatchedCount:         3,
			expectedUnmatchedSystemCount: 1,
			expectedUnmatchedBankCount:   1,
			expectedDiscrepancy:          newDecimalFromString("500.00"),
		},
		{
			name:      "Unmatched Bank Tx in a Group with No System Tx",
			systemTxs: []domain.SystemTransaction{},
			bankTxs: []domain.BankTransaction{
				{ID: "B6-ISOLATED", Amount: newDecimalFromString("-50.00"), Date: newDate(6)},
			},
			expectedMatchedCount:         0,
			expectedUnmatchedSystemCount: 0,
			expectedUnmatchedBankCount:   1,
			expectedDiscrepancy:          newDecimalFromString("0"),
		},
		{
			name: "Coverage for bankTxUsed check",
			systemTxs: []domain.SystemTransaction{
				{ID: "S-COVERAGE-1", Amount: newDecimalFromString("100"), Type: domain.Debit, TransactionTime: newDate(10)},
				{ID: "S-COVERAGE-2", Amount: newDecimalFromString("200"), Type: domain.Debit, TransactionTime: newDate(10)},
			},
			bankTxs: []domain.BankTransaction{
				{ID: "B-COVERAGE", Amount: newDecimalFromString("-101"), Date: newDate(10)},
			},
			expectedMatchedCount:         1,
			expectedUnmatchedSystemCount: 1,
			expectedUnmatchedBankCount:   0,
			expectedDiscrepancy:          newDecimalFromString("1"),
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
