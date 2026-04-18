# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically-linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# ── Run stage ─────────────────────────────────────────────────────────────────
FROM alpine:3.20

# ca-certificates needed for TLS connections to Azure SQL / Blob
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

# Override the default "localhost:8000" so the port is reachable inside Docker
ENV HTTP_ADDRESS=0.0.0.0:8000

EXPOSE 8000

ENTRYPOINT ["./server"]
