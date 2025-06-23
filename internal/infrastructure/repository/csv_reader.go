package repository

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nmmugia/reconciliation-service/internal/domain"

	"github.com/shopspring/decimal"
)

type CsvLedgerReader struct{}

func NewCsvLedgerReader() *CsvLedgerReader {
	return &CsvLedgerReader{}
}

func (r *CsvLedgerReader) ReadSystemTransactions(filePath string, startDate, endDate time.Time) ([]domain.SystemTransaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open system transaction file '%s': %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("could not read header from '%s': %w", filePath, err)
	}

	var transactions []domain.SystemTransaction
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading record from '%s': %w", filePath, err)
		}

		txTime, err := time.Parse(time.RFC3339, record[3])
		if err != nil {
			continue
		}

		if txTime.Before(startDate) || txTime.After(endDate.Add(24*time.Hour-time.Nanosecond)) {
			continue
		}

		amount, err := decimal.NewFromString(record[1])
		if err != nil {
			continue
		}

		transactions = append(transactions, domain.SystemTransaction{
			ID:              record[0],
			Amount:          amount,
			Type:            domain.TransactionType(record[2]),
			TransactionTime: txTime,
		})
	}
	return transactions, nil
}

func (r *CsvLedgerReader) ReadBankTransactions(filePaths []string, startDate, endDate time.Time) ([]domain.BankTransaction, error) {
	var allBankTransactions []domain.BankTransaction
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(filePaths))

	for _, path := range filePaths {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			bankName := filepath.Base(filePath)
			transactions, err := r.parseSingleBankStatement(filePath, bankName, startDate, endDate)
			if err != nil {
				errChan <- err
				return
			}
			mu.Lock()
			allBankTransactions = append(allBankTransactions, transactions...)
			mu.Unlock()
		}(path)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return allBankTransactions, nil
}

func (r *CsvLedgerReader) parseSingleBankStatement(filePath, bankName string, startDate, endDate time.Time) ([]domain.BankTransaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open bank statement file '%s': %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("could not read header from '%s': %w", filePath, err)
	}

	var transactions []domain.BankTransaction
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading record from '%s': %w", filePath, err)
		}

		date, err := time.Parse("2006-01-02", record[2])
		if err != nil {
			continue
		}

		if date.Before(startDate) || date.After(endDate) {
			continue
		}

		amount, err := decimal.NewFromString(record[1])
		if err != nil {
			continue
		}

		transactions = append(transactions, domain.BankTransaction{
			ID:       record[0],
			Amount:   amount,
			Date:     date,
			BankName: bankName,
		})
	}
	return transactions, nil
}
