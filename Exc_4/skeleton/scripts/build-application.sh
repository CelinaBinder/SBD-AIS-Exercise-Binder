#!/bin/sh
set -e

echo "Building Go application..."

# Optional: ensure dependencies are up to date
go mod tidy

# Build a fully static binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ordersystem .

echo "Build complete: ./ordersystem"
