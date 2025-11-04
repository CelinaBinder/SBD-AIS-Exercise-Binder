# Step 1: Build the Go binary
FROM golang:1.25 AS builder

WORKDIR /app

# Copy go.mod and go.sum first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the source
COPY . .

# Build the Go application
RUN chmod +x scripts/build-application.sh && ./scripts/build-application.sh

# Step 2: Create minimal runtime image
FROM alpine:latest

# Install certificates for HTTPS
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy compiled binary from builder stage
COPY --from=builder /app/ordersystem .

# Expose the HTTP port
EXPOSE 3000

# Environment variables (defaults; can be overridden in docker-compose)
ENV DB_HOST=ordersystem-db \
    DB_PORT=5432 \
    DB_USER=docker \
    DB_PASSWORD=docker \
    DB_NAME=order

# Command to start your binary
CMD ["./ordersystem"]
