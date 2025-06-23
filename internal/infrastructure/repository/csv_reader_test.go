package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempCsv(t *testing.T, content string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "test_*.csv")
	require.NoError(t, err, "Failed to create temp file")
	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err, "Failed to write to temp file")
	require.NoError(t, tmpfile.Close(), "Failed to close temp file")
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })
	return tmpfile.Name()
}

func TestCsvLedgerReader_ReadSystemTransactions(t *testing.T) {
	reader := NewCsvLedgerReader()
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		content := `trxID,amount,type,transactionTime
sys001,100,CREDIT,2023-01-15T10:00:00Z
sys002,200,DEBIT,2023-02-01T10:00:00Z`
		filePath := createTempCsv(t, content)
		txs, err := reader.ReadSystemTransactions(filePath, startDate, endDate)
		require.NoError(t, err)
		assert.Len(t, txs, 1)
		assert.Equal(t, "sys001", txs[0].ID)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := reader.ReadSystemTransactions("non_existent_file.csv", startDate, endDate)
		assert.Error(t, err)
	})

	t.Run("bad header read", func(t *testing.T) {
		filePath := createTempCsv(t, "")
		_, err := reader.ReadSystemTransactions(filePath, startDate, endDate)
		assert.Error(t, err)
	})

	t.Run("bad record read", func(t *testing.T) {
		content := `trxID,amount,type,transactionTime
"sys001,100,CREDIT,2023-01-15T10:00:00Z`
		filePath := createTempCsv(t, content)
		_, err := reader.ReadSystemTransactions(filePath, startDate, endDate)
		assert.Error(t, err)
	})

	t.Run("malformed date", func(t *testing.T) {
		content := `trxID,amount,type,transactionTime
sys001,100,CREDIT,2023/01/15`
		filePath := createTempCsv(t, content)
		txs, err := reader.ReadSystemTransactions(filePath, startDate, endDate)
		require.NoError(t, err)
		assert.Len(t, txs, 0)
	})

	t.Run("malformed amount", func(t *testing.T) {
		content := `trxID,amount,type,transactionTime
sys001,abc,CREDIT,2023-01-15T10:00:00Z`
		filePath := createTempCsv(t, content)
		txs, err := reader.ReadSystemTransactions(filePath, startDate, endDate)
		require.NoError(t, err)
		assert.Len(t, txs, 0)
	})
}

func TestCsvLedgerReader_ReadBankTransactions(t *testing.T) {
	reader := NewCsvLedgerReader()
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	t.Run("success with multiple files", func(t *testing.T) {
		content1 := `unique_identifier,amount,date
bank001,100,2023-01-15`
		content2 := `unique_identifier,amount,date
bank002,200,2023-01-20
bank003,300,2023-02-10`
		file1 := createTempCsv(t, content1)
		file2 := createTempCsv(t, content2)

		txs, err := reader.ReadBankTransactions([]string{file1, file2}, startDate, endDate)
		require.NoError(t, err)
		assert.Len(t, txs, 2)
		var bankNames []string
		for _, tx := range txs {
			bankNames = append(bankNames, tx.BankName)
		}
		assert.Contains(t, bankNames, filepath.Base(file1))
		assert.Contains(t, bankNames, filepath.Base(file2))
	})

	t.Run("one file not found", func(t *testing.T) {
		content1 := `unique_identifier,amount,date
bank001,100,2023-01-15`
		file1 := createTempCsv(t, content1)
		_, err := reader.ReadBankTransactions([]string{file1, "non_existent_file.csv"}, startDate, endDate)
		assert.Error(t, err)
	})

	t.Run("one file has bad header", func(t *testing.T) {
		file1 := createTempCsv(t, "")
		_, err := reader.ReadBankTransactions([]string{file1}, startDate, endDate)
		assert.Error(t, err)
	})

	t.Run("one file has bad record", func(t *testing.T) {
		content := `unique_identifier,amount,date
"bank001,100,2023-01-15`
		file1 := createTempCsv(t, content)
		_, err := reader.ReadBankTransactions([]string{file1}, startDate, endDate)
		assert.Error(t, err)
	})

	t.Run("malformed date", func(t *testing.T) {
		content := `unique_identifier,amount,date
bank001,100,not-a-date`
		filePath := createTempCsv(t, content)
		txs, err := reader.ReadBankTransactions([]string{filePath}, startDate, endDate)
		require.NoError(t, err)
		assert.Empty(t, txs)
	})

	t.Run("malformed amount", func(t *testing.T) {
		content := `unique_identifier,amount,date
bank001,not-an-amount,2023-01-15`
		filePath := createTempCsv(t, content)
		txs, err := reader.ReadBankTransactions([]string{filePath}, startDate, endDate)
		require.NoError(t, err)
		assert.Empty(t, txs)
	})
}
