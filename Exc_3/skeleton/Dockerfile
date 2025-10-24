  # todo
# Step 1: Build it
FROM golang:1.25 AS builder

WORKDIR /app

# Copy everything into /app
COPY . .

# Build the Go application
RUN chmod +x scripts/build-application.sh && ./scripts/build-application.sh

# Step 2: Running it
FROM alpine:latest

# Install certificates for HTTPS support
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/ordersystem .

# HTTP port
EXPOSE 3000

# environment variables
ENV DB_HOST=host.docker.internal
ENV DB_PORT=5432
ENV DB_USER=docker
ENV DB_PASSWORD=docker
ENV DB_NAME=order

# Start the application
CMD ["./ordersystem"]
