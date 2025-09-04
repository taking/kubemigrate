# Build stage
FROM golang:1.24-alpine3.22 AS builder

LABEL authors="taking"
LABEL description="Velero API Server - Multi-cluster backup and restore management"

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o velero-api ./cmd/main.go

# Runtime stage
FROM alpine:3.22 AS runtime

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/velero-api .

# Create swagger directories
RUN mkdir -p ./docs/swagger

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:9091/api/v1/health || exit 1

# Expose port
EXPOSE 9091

# Run the application
CMD ["./velero-api"]