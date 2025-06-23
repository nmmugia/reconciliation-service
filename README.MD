# Transaction Reconciliation Service

## 1. Project Description

This service provides a robust, high-performance solution for reconciling financial transactions between an internal system and external bank statements. It is engineered to process large volumes of data with speed and accuracy—identifying matched transactions, amount discrepancies, and unmatched items requiring manual review.

The system follows **Clean Architecture principles** using **Go**, ensuring a clear separation of concerns that makes the application maintainable, scalable, and testable. It leverages Go’s **concurrency** capabilities to optimize performance for I/O-bound operations like file reading.

---

## 2. The Reconciliation Algorithm: *Best Fit with Threshold*

To achieve logical and high-accuracy results, the service uses a **Best Fit with Threshold** algorithm—in which will consider the matching transaction based on maximum set treshold (currently: 1.000), and if the transaction discrepancy is more than (let's say) 1.000, then the transaction will be considered as unmatched transaction.


## 3. Installation and Execution

### ✅ Prerequisites

- [Go](https://golang.org/doc/install) (1.23.9+)
- [Docker](https://www.docker.com/products/docker-desktop)

---

### ▶ Option 1: Run with Go

```bash
git clone https://github.com/nmmugia/reconciliation-service.git
cd reconciliation-service

go mod tidy
```

**Run the application:**

```bash
go run ./cmd/reconciler/ \
  -sys="/path/to/system_transactions.csv" \
  -bank="/path/to/bank1.csv,/path/to/bank2.csv" \
  -start="2025-06-01" \
  -end="2025-06-30"
```

---

### ▶ Option 2: Run with Docker *(Recommended)*

**Build the production image:**

```bash
docker build -t reconciler-app:latest -f Dockerfile .
```

**Run the container:**

```bash
docker run --rm \
  -v /path/to/your/local/data:/data \
  reconciler-app:latest \
  -sys="/data/system_transactions.csv" \
  -bank="/data/bank_statement_amartha.csv" \
  -start="2025-06-23" \
  -end="2025-06-24"
```

---

## 5. Testing

### ✅ Unit Tests

Run unit tests and view coverage:

```bash
go test -v --cover ./internal/...
```

![unit test result](https://github.com/nmmugia/reconciliation-service/blob/master/sample/ss2.png?raw=true)

---

### ✅ Integration Test

**Build the application and test images:**

```bash
docker build -t reconciler-app:latest -f Dockerfile .
docker build -t reconciler-test:latest -f Dockerfile.test .
```

**Run integration test:**

```bash
docker run --rm reconciler-test:latest
```

> Success message:

![integration test result](https://github.com/nmmugia/reconciliation-service/blob/master/sample/ss.png?raw=true)

---


