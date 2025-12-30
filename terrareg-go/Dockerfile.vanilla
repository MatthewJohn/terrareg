# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /bin/terrareg cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 terrareg && \
    adduser -D -u 1000 -G terrareg terrareg

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/terrareg /app/terrareg

# Copy static files and templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/templates /app/templates

# Create data directories
RUN mkdir -p /app/data /app/data/upload && \
    chown -R terrareg:terrareg /app

# Switch to non-root user
USER terrareg

# Expose port
EXPOSE 5000

# Set environment variables
ENV LISTEN_PORT=5000
ENV DATA_DIRECTORY=/app/data
ENV UPLOAD_DIRECTORY=/app/data/upload

# Run the application
CMD ["/app/terrareg"]
