FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /reconciler -ldflags="-w -s" ./cmd/reconciler

FROM scratch

COPY --from=builder /reconciler /reconciler

ENTRYPOINT ["/reconciler"]