#!/bin/sh
set -e

echo "--- Starting Integration Test ---"

SYS_TX_FILE="/data/system.csv"
BANK_TX_FILE="/data/bank.csv"

if [ ! -f "$SYS_TX_FILE" ] || [ ! -f "$BANK_TX_FILE" ]; then
    echo "Error: Test data not found. Ensure sample_data is mounted to /data." >&2
    exit 1
fi

echo "Running reconciler..."
output=$(/reconciler \
    -sys="$SYS_TX_FILE" \
    -bank="$BANK_TX_FILE" \
    -start="2025-06-23" \
    -end="2025-06-24")

check_output() {
    local expected="$1"
    echo -n "Checking for: '$expected' ... "
    if echo "$output" | grep -qF "$expected"; then
        echo "[PASS]"
    else
        echo "[FAIL]"
        echo "Error: Expected output not found." >&2
        echo "--- Full Application Output ---" >&2
        echo "$output" >&2
        echo "-----------------------------" >&2
        exit 1
    fi
}

echo "--- Verifying Reconciliation Summary ---"

check_output "Total System Transactions Processed: 4"
check_output "Total Bank Transactions Processed:   4"
check_output "Matched Transactions:                3"
check_output "Unmatched System Transactions:       1"
check_output "Unmatched Bank Transactions:         1"

check_output "Total Amount Discrepancy:            500.00"

echo "--- Verifying Unmatched Transaction Details ---"

check_output "ID: SYS-004"
check_output "ID: BNK-FEE-X"

echo ""
echo "--- Integration Test Successful: Application behaves as expected. ---"
exit 0