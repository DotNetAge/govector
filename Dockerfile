# =============================================================================
# GoVector Docker Image — Lightweight vector database HTTP API server
# =============================================================================
#
# Build:
#   docker build -t govector:latest .
#
# Run:
#   docker run -d \
#     --name govector \
#     -p 18080:18080 \
#     -v govector-data:/data \
#     govector:latest
#
# With custom port and DB path:
#   docker run -d \
#     --name govector \
#     -p 8080:18080 \
#     -v ./mydb:/data \
#     -e GOVECTOR_PORT=8080 \
#     -e GOVECTOR_DB=/data/mydb.db \
#     govector:latest
# =============================================================================

FROM golang:alpine AS builder

WORKDIR /src

# Cache dependency downloads (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Build static binary (CGO_ENABLED=0 = fully static, no libc dependency)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /go/bin/govector ./cmd/govector

# ---- Runtime image ----
FROM alpine:3.19

LABEL maintainer="DotNetAge <ray@dotnetage.com>"
LABEL org.opencontainers.image.source="https://github.com/DotNetAge/govector"
LABEL org.opencontainers.image.description="GoVector - High-performance embedded vector database"

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g "" govector

USER govector
WORKDIR /data

COPY --from=builder /go/bin/govector /usr/local/bin/govector

# Default environment variables
ENV GOVECTOR_PORT=18080
ENV GOVECTOR_DB=/data/govector.db

EXPOSE 18080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:${GOVECTOR_PORT:-18080}/collections || exit 1

ENTRYPOINT ["sh", "-c", "govector serve -port=${GOVECTOR_PORT:-18080} -db=${GOVECTOR_DB:-/data/govector.db}"]
