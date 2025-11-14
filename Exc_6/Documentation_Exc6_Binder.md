# 🐳 Docker Setup & S3 Integration — Order Service System

## 📘 Project Overview
This documentation outlines the step-by-step process of building, debugging, and running the **Order Service system** using Docker Compose, including **backend, frontend, PostgreSQL, Traefik reverse proxy**, and **object storage via MinIO/S3**.

## ⚙️ File Overview

### 1. Dockerfile
```dockerfile
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN sh /app/scripts/build-application.sh

FROM alpine AS run
WORKDIR /
COPY --from=builder /app/ordersystem /app/ordersystem
EXPOSE 3000
CMD ["/app/ordersystem"]
```

### 2. Build Script
```bash
#!/bin/sh
set -e
cd /app || exit 1
go mod tidy
go mod download
CGO_ENABLED=0 GOOS=linux go build -o /app/ordersystem
```

### 3. docker-compose.yml
```yaml
networks:
  web:
    name: web
  intercom:
    name: intercom

volumes:
  order_pg_vol:
  order_minio_vol:

services:
  traefik:
    image: traefik:v3.5.2
    container_name: traefik
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - //var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - web

  orderservice:
    container_name: orderservice
    build: .
    command: ["/app/ordersystem"]
    environment:
      - POSTGRES_DB=order
      - POSTGRES_USER=docker
      - POSTGRES_PASSWORD=docker
      - POSTGRES_TCP_PORT=5555
      - DB_HOST=order-postgres
      - S3_ACCESS_KEY_ID=root
      - S3_SECRET_ACCESS_KEY=verysecret
      - S3_ENDPOINT=localhost:8500
    depends_on:
      order-postgres:
        condition: service_healthy
    networks:
      - intercom
      - web
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.orderservice.rule=Host(`orders.localhost`)"
      - "traefik.http.routers.orderservice.entrypoints=web"
      - "traefik.http.services.orderservice.loadbalancer.server.port=3000"

  sws:
    image: joseluisq/static-web-server:latest
    container_name: sws
    volumes:
      - ./frontend:/public:ro
    environment:
      - SERVER_PORT=80
      - SERVER_ROOT=/public
    networks:
      - web
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.sws.rule=Host(`localhost`)"
      - "traefik.http.routers.sws.entrypoints=web"
      - "traefik.http.services.sws.loadbalancer.server.port=80"

  order-postgres:
    image: postgres:latest
    restart: always
    networks:
      - intercom
    volumes:
      - order_pg_vol:/var/lib/postgresql
    ports:
      - "5555:5555"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U docker -d order"]
      interval: 5s
      timeout: 5s
      retries: 10
    environment:
      - POSTGRES_DB=order
      - POSTGRES_USER=docker
      - POSTGRES_PASSWORD=docker
      - POSTGRES_TCP_PORT=5555

  order-minio:
    image: minio/minio:latest
    container_name: order-minio
    volumes:
      - order_minio_vol:/data
    environment:
      - MINIO_ROOT_USER=root
      - MINIO_ROOT_PASSWORD=verysecret
    command: server /data --console-address ":42143"
    ports:
      - "8500:8500"
      - "42143:42143"
    networks:
      - web
```

## 🧾 Troubleshooting & Fixes
1. **CRLF vs LF line endings** → Fixed in `build-application.sh`.
2. **Go modules missing** → `go mod tidy` and `go mod download` added.
3. **Postgres not ready** → Healthcheck added and backend depends_on `service_healthy`.
4. **Postgres 18 directory changes** → Removed old volumes.
5. **Traefik routing issues** → Correct networks and labels.

## ☁️ S3 / MinIO Integration
**Access via IDE plugin (Remote File Systems / AWS S3 plugin):**

- **Host:** `localhost`
- **Port:** `8500`
- **Authentication type:** Access Key / Secret Key
- **Access Key:** `root`
- **Secret Key:** `verysecret`
- **Path style access:** enabled
- **Use SSL:** disabled

Steps in IDE:
1. Open Remote File Systems plugin.
2. Add connection → AWS S3.
3. Enter host, port, credentials, disable SSL.
4. Test connection and connect.

**You can now browse MinIO buckets directly in your IDE.**

## 🖥 Final URLs
| Service              | URL                                                                 | Description                  |
|---------------------|---------------------------------------------------------------------|------------------------------|
| Backend (OpenAPI)   | [http://orders.localhost/openapi/index.html#/Menu/get_api_menu](http://orders.localhost/openapi/index.html#/Menu/get_api_menu) | Backend API docs             |
| Frontend (SWS)      | [http://localhost/](http://localhost/)                               | Static frontend UI           |
| MinIO Web UI        | [http://localhost:42143](http://localhost:42143)                     | Object storage management    |
| Traefik Dashboard   | [http://localhost:8080](http://localhost:8080)                       | Reverse proxy dashboard      |

## 📝 Order Model (Go)
```go
package model

import (
	"fmt"
	"time"
)

const orderFilename = "order_%d.md"
const markdownTemplate = `# Order: %d

| Created At       | Drink ID | Amount |
|-----------------|----------|--------|
| %s | %d        | %d     |
`

type Order struct {
	Base
	Amount uint64 `json:"amount"`
	DrinkID uint  `json:"drink_id" gorm:"not null"`
	Drink   Drink `json:"drink"`
}

func (o *Order) ToMarkdown() string {
	return fmt.Sprintf(markdownTemplate, o.ID, o.CreatedAt.Format(time.RFC1123), o.DrinkID, o.Amount)
}

func (o *Order) GetFilename() string {
	return fmt.Sprintf(orderFilename, o.ID)
}
```

---

This setup allows full **backend + frontend + DB + object storage** integration using Docker, MinIO, and IDE S3 plugin access.