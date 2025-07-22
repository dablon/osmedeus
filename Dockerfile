# Multi-stage Dockerfile for Osmedeus
# Stage 1: Build stage with Go compilation
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    sqlite-dev \
    ca-certificates \
    tzdata \
    make \
    pkgconfig

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o osmedeus .

# Stage 2: Runtime stage with Alpine base
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite \
    curl \
    bash \
    git \
    nmap \
    masscan \
    wget \
    unzip \
    && rm -rf /var/cache/apk/*

# Create osmedeus user
RUN addgroup -g 1000 osmedeus && \
    adduser -D -s /bin/bash -u 1000 -G osmedeus osmedeus

# Create osmedeus home directory structure
RUN mkdir -p /home/osmedeus && \
    chown -R osmedeus:osmedeus /home/osmedeus

# Switch to osmedeus user for base installation
USER osmedeus
WORKDIR /home/osmedeus

# Install osmedeus-base
RUN bash -c "$(curl -fsSL https://raw.githubusercontent.com/osmedeus/osmedeus-base/master/install.sh)" || true

# Switch back to root to copy binary and set permissions
USER root

# Copy binary from builder stage
COPY --from=builder /app/osmedeus /usr/local/bin/osmedeus

# Create app directories
RUN mkdir -p \
    /app/data \
    /app/logs \
    /app/workspaces \
    /app/config \
    && chown -R osmedeus:osmedeus /app

# Copy source code for reference (optional)
COPY --chown=osmedeus:osmedeus . /app/src/

# Set working directory
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=15s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Switch to osmedeus user
USER osmedeus

# Expose ports
EXPOSE 8000 8080

# Set environment variables
ENV GO_ENV=production
ENV REDIS_URL=redis://redis:6379
ENV HOME=/home/osmedeus
ENV OSMEDEUS_BASE_PATH=/home/osmedeus/osmedeus-base

# Default command
CMD ["osmedeus", "server"]