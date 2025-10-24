#!/bin/bash
set -e

# Remove old DB container & volume
docker rm -f orders-db || true
docker volume rm pgdata || true
docker volume create pgdata

# Start PostgreSQL container
docker run -d \
  --name orders-db \
  -e POSTGRES_DB=order \
  -e POSTGRES_USER=docker \
  -e POSTGRES_PASSWORD=docker \
  -v pgdata:/var/lib/postgresql/18/docker \
  -p 5432:5432 \
  postgres:18
