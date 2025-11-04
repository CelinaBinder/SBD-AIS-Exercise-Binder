# Exercise 4 – Software Architecture for Big Data: Dockerizing PostgreSQL and Order Service

## Objective

Set up a Dockerized environment for the `orderservice` application with a PostgreSQL database. The goal was to:

* Use Docker Compose to orchestrate services.
* Ensure network communication between database and service.
* Build a fully functional Go binary for deployment.
* Expose correct ports and environment variables.
* Enable persistent storage for PostgreSQL.

---

## Project Structure

```
skeleton/
├─ Dockerfile
├─ docker-compose.yml
├─ go.mod
├─ go.sum
├─ main.go
├─ scripts/
│  └─ build-application.sh
├─ ordersystem/ (optional, depending on your Go files)
```

---

## Step 1 – Dockerfile

We used a **multi-stage build** for the Go application to create a minimal runtime image:

```dockerfile
# Stage 1: Build the Go binary
FROM golang:1.25 AS builder

WORKDIR /app

# Copy module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Make build script executable and run it
RUN chmod +x scripts/build-application.sh && ./scripts/build-application.sh

# Stage 2: Minimal runtime image
FROM alpine:latest

# Install certificates for HTTPS support
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/ordersystem .

# Expose HTTP port
EXPOSE 3000

# Start the application
CMD ["./ordersystem"]
```

**Notes:**

* Multi-stage build reduces runtime image size.
* Alpine is used as runtime but a fully static binary ensures compatibility.
* The binary `ordersystem` is built with `CGO_ENABLED=0` to avoid glibc dependencies.

---

## Step 2 – Build Script

`scripts/build-application.sh` builds the Go application:

```bash
#!/bin/sh
set -e

echo "🔧 Building statically linked Go binary..."

# Build a fully static binary for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ordersystem .

echo "✅ Build complete: ./ordersystem"
```

**Notes:**

* `CGO_ENABLED=0` ensures a static binary for Alpine runtime.
* The binary is output directly to `/app/ordersystem`.

---

## Step 3 – Docker Compose File

`docker-compose.yml` orchestrates the services:

```yaml
version: "3.9"

services:
  ordersystem-db:
    image: postgres:18
    container_name: ordersystem-db
    environment:
      POSTGRES_DB: order
      POSTGRES_USER: docker
      POSTGRES_PASSWORD: docker
      PGDATA: /var/lib/postgresql/data
    volumes:
      - pgdata_sbd:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    networks:
      - sbdnetwork

  orderservice:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: orderservice
    environment:
      POSTGRES_DB: order
      POSTGRES_USER: docker
      POSTGRES_PASSWORD: docker
      POSTGRES_TCP_PORT: 5432
      DB_HOST: ordersystem-db
    ports:
      - "3000:3000"
    depends_on:
      - ordersystem-db
    networks:
      - sbdnetwork

volumes:
  pgdata_sbd:

networks:
  sbdnetwork:
```

**Notes:**

* The `orderservice` depends on `ordersystem-db


### Stop everything
- to stop everything before closing: docker compose down
- check it actually stopped: docker ps
- 
### Start everything
- start all containers: docker compose up -d
- again verify its running: docker ps
