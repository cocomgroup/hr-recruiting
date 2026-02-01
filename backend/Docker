# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy all source code
COPY . .

# Tidy and download dependencies
RUN go mod tidy && go mod download

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /app/server \
    ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/app/server"]