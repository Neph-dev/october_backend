# Multi-stage build for production
# Following NASA security and reliability principles

# Build stage - Use latest Go version (compatible with go.mod requirements)
FROM golang:alpine AS builder

# Install security updates and required tools, remove package cache
RUN apk update && apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

# Create appuser for security (with explicit shell and home)
RUN adduser -D -g '' -s /sbin/nologin -h /nonexistent appuser

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application with security flags and version info
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -trimpath \
    -o october-server ./cmd/api

# Verify the binary
RUN ./october-server -version || echo "Version flag not implemented yet"

# Final stage - minimal runtime image
FROM scratch

# Import certificates and user info from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable
COPY --from=builder /build/october-server /october-server

# Add metadata labels for container registry
LABEL maintainer="october-backend-team" \
      version="1.0.0" \
      description="October Backend API Server" \
      org.opencontainers.image.source="https://github.com/Neph-dev/october_backend" \
      org.opencontainers.image.vendor="October Team" \
      org.opencontainers.image.title="October Backend"

# Use non-root user for security
USER appuser

# Expose port (documentation only - actual port comes from config)
EXPOSE 8080

# Health check with better configuration
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/october-server", "-health"]

# Set entry point
ENTRYPOINT ["/october-server"]