#!/bin/bash
set -e

# Remove old container if exists
docker rm -f orderservice || true

# Build the image, specifying Dockerfile path and build context
docker build -f "../Dockerfile" -t orderservice:latest ..

# Run the container with given credentials
docker run -d \
  --name orderservice \
  -e DB_HOST=orders-db \
  -e DB_PORT=5432 \
  -e DB_USER=docker \
  -e DB_PASSWORD=docker \
  -e DB_NAME=order \
  -p 3000:3000 \
  orderservice:latest
