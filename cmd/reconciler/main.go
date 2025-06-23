package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nmmugia/reconciliation-service/internal/domain"
	"github.com/nmmugia/reconciliation-service/internal/infrastructure/repository"
	"github.com/nmmugia/reconciliation-service/internal/service"
	"github.com/nmmugia/reconciliation-service/internal/usecase"
)

func main() {

	processStartTime := time.Now()

	sysTxPath := flag.String("sys", "", "Path to system transactions CSV. (Required)")
	bankStatementPaths := flag.String("bank", "", "Comma-separated paths to bank statement CSVs. (Required)")
	startDateStr := flag.String("start", "", "Start date for reconciliation (YYYY-MM-DD). (Required)")
	endDateStr := flag.String("end", "", "End date for reconciliation (YYYY-MM-DD). (Required)")
	flag.Parse()

	if *sysTxPath == "" || *bankStatementPaths == "" || *startDateStr == "" || *endDateStr == "" {
		fmt.Println("Error: All flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	startDate, err := time.Parse("2006-01-02", *startDateStr)
	if err != nil {
		log.Fatalf("Invalid start date format: %v. Please use YYYY-MM-DD.", err)
	}
	endDate, err := time.Parse("2006-01-02", *endDateStr)
	if err != nil {
		log.Fatalf("Invalid end date format: %v. Please use YYYY-MM-DD.", err)
	}

	csvReader := repository.NewCsvLedgerReader()
	recoEngine := service.NewReconciliationEngine()
	reconciler := usecase.NewReconciliationUsecase(csvReader, csvReader, recoEngine)

	log.Println("Starting reconciliation process...")
	summary, err := reconciler.PerformReconciliation(*sysTxPath, strings.Split(*bankStatementPaths, ","), startDate, endDate)
	if err != nil {
		log.Fatalf("Reconciliation failed: %v", err)
	}
	log.Println("Reconciliation process completed successfully.")
	summary.ProcessingDurationSeconds = time.Since(processStartTime).Seconds()

	printSummary(summary)
}

func printSummary(summary *domain.ReconciliationSummary) {
	fmt.Println("\n--- Reconciliation Report ---")
	fmt.Printf("Processing Time: %.2f seconds\n\n", summary.ProcessingDurationSeconds)
	fmt.Println("[Summary]")
	fmt.Printf("Total System Transactions Processed: %d\n", summary.TotalSystemTransactions)
	fmt.Printf("Total Bank Transactions Processed:   %d\n", summary.TotalBankTransactions)
	fmt.Printf("Matched Transactions:                %d\n", summary.MatchedTransactions)
	fmt.Printf("Unmatched System Transactions:       %d\n", len(summary.UnmatchedSystemTransactions))
	unmatchedBankCount := 0
	for _, txs := range summary.UnmatchedBankTransactions {
		unmatchedBankCount += len(txs)
	}
	fmt.Printf("Unmatched Bank Transactions:         %d\n", unmatchedBankCount)
	fmt.Printf("Total Amount Discrepancy:            %s\n", summary.AmountDiscrepancyTotal.StringFixed(2))

	if len(summary.UnmatchedSystemTransactions) > 0 {
		fmt.Println("\n[Unmatched System Transactions]")
		for _, tx := range summary.UnmatchedSystemTransactions {
			fmt.Printf("- ID: %s, Amount: %s, Type: %s, Time: %s\n",
				tx.ID, tx.Amount.StringFixed(2), tx.Type, tx.TransactionTime.Format(time.RFC3339))
		}
	}

	if len(summary.UnmatchedBankTransactions) > 0 {
		fmt.Println("\n[Unmatched Bank Transactions]")
		for bank, txs := range summary.UnmatchedBankTransactions {
			fmt.Printf("  Bank: %s\n", bank)
			for _, tx := range txs {
				txType := "CREDIT"
				if tx.Amount.IsNegative() {
					txType = "DEBIT"
				}
				fmt.Printf("  - ID: %s, Amount: %s (%s), Date: %s\n",
					tx.ID, tx.Amount.Abs().StringFixed(2), txType, tx.Date.Format("2006-01-02"))
			}
		}
	}
	fmt.Println("\n--- End of Report ---")
}
