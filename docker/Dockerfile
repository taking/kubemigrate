FROM golang:1.24-alpine3.22 AS builder

LABEL authors="taking"
LABEL description="KubeMigrate API Server - Multi-cluster backup and restore management"

RUN apk add --no-cache git ca-certificates tzdata upx

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN echo "🔧 Building for GOOS=$TARGETOS GOARCH=$TARGETARCH" && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -a -installsuffix cgo \
    -ldflags '-s -w -extldflags "-static"' \
    -trimpath \
    -buildvcs=false \
    -o kubemigrate-api ./cmd/main.go

# UPX 압축 최적화 (linux/amd64인 경우에만)
RUN if [ "$TARGETOS" = "linux" ] && [ "$TARGETARCH" = "amd64" ]; then \
        echo "🗜️  Compressing binary with UPX for linux/amd64..." && \
        upx --best --lzma kubemigrate-api -o kubemigrate-api-compressed && \
        mv kubemigrate-api-compressed kubemigrate-api; \
    else \
        echo "⏭️  Skipping UPX compression for $TARGETOS/$TARGETARCH"; \
    fi

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary and Swagger docs
COPY --from=builder /app/kubemigrate-api /kubemigrate-api
COPY --from=builder /app/docs/swagger /docs/swagger

# Metadata
LABEL org.opencontainers.image.source="https://github.com/taking/kubemigrate"

# Expose port
EXPOSE 9091

# Run as non-root
USER 1001:1001

# Default command
ENTRYPOINT ["/kubemigrate-api"]
