backend: http://orders.localhost/openapi/index.html#/Menu/get_api_menu

frontend: http://localhost/

# 🐳 Docker Setup & Troubleshooting Log — Order Service System

## 📘 Project Overview
This documentation outlines the step-by-step process of building, debugging, and successfully running the **Order Service system** consisting of:
- A **Go backend** (`orderservice`)
- A **PostgreSQL database**
- A **Static frontend server**
- A **Traefik reverse proxy**

All services are orchestrated using **Docker Compose**.

---

## 🧩 Initial Setup
---

````markdown
# 🐳 Docker Setup, Explanation & Troubleshooting Log — Order Service System

## 📘 Project Overview
This documentation explains the **Order Service System** — how it was built, debugged, and run successfully using Docker and Docker Compose.  
It combines:
- A **Go backend** (`orderservice`)
- A **PostgreSQL database**
- A **Static frontend server**
- A **Traefik reverse proxy**

All services are orchestrated using **Docker Compose**.

---

## ⚙️ File Overview

### 🧱 1. Dockerfile

#### Full File
```dockerfile
FROM golang:1.25 AS builder
WORKDIR /app

# Copy only go.mod and go.sum first (better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the Go binary
RUN sh /app/scripts/build-application.sh


FROM alpine AS run
WORKDIR /
COPY --from=builder /app/ordersystem /app/ordersystem
# EXPOSE doesn't actually do anything!
EXPOSE 3000
CMD ["/app/ordersystem"]
````

#### 📖 Explanation (Line by Line)

1. `FROM golang:1.25 AS builder`
   → Use the Go 1.25 image to compile the application. Creates a **build stage**.

2. `WORKDIR /app`
   → Sets `/app` as the working directory.

3. `COPY go.mod go.sum ./`
   → Copies dependency files for efficient caching.

4. `RUN go mod download`
   → Downloads all Go module dependencies.

5. `COPY . .`
   → Copies the rest of the application source code.

6. `RUN sh /app/scripts/build-application.sh`
   → Runs the custom build script (next section) to compile the binary.

7. `FROM alpine AS run`
   → Starts a new, smaller runtime stage (minimal base image).

8. `WORKDIR /`
   → Sets the working directory at root.

9. `COPY --from=builder /app/ordersystem /app/ordersystem`
   → Copies the compiled binary from the builder stage.

10. `EXPOSE 3000`
    → Declares that the app listens on port 3000 (informational only).

11. `CMD ["/app/ordersystem"]`
    → Defines the default command to run when the container starts.

---

### ⚙️ 2. Build Script

#### Full File

```bash
#!/bin/sh
# Exit immediately if any command fails
set -e

# Change to app directory
cd /app || exit 1

# Download Go modules
go mod download

# Build the Go binary for Linux
CGO_ENABLED=0 GOOS=linux go build -o /app/ordersystem
```

#### 📖 Explanation (Line by Line)

1. `#!/bin/sh`
   → Tells Docker to execute this script using the `sh` shell.

2. `set -e`
   → Stop immediately if any command fails.

3. `cd /app || exit 1`
   → Navigate into the `/app` directory (exit if it doesn’t exist).

4. `go mod download`
   → Fetches all module dependencies required by the Go app.

5. `CGO_ENABLED=0 GOOS=linux go build -o /app/ordersystem`
   → Builds a statically linked Linux binary named `ordersystem`.

---

### 🧩 3. docker-compose.yml

#### Full File

```yaml
networks:
  web:
    name: web
  intercom:
    name: intercom

volumes:
  order_pg_vol:

services:
  traefik:
    image: "traefik:v3.5.2"
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
    command: [ "/app/ordersystem" ]
    environment:
      - POSTGRES_DB=order
      - POSTGRES_USER=docker
      - POSTGRES_PASSWORD=docker
      - POSTGRES_TCP_PORT=5432
      - DB_HOST=order-postgres
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
      - "5432:5432"
    healthcheck:
      test: [ "CMD-SHELL", "pg_isready -U docker -d order" ]
      interval: 5s
      timeout: 5s
      retries: 10
    environment:
      - POSTGRES_DB=order
      - POSTGRES_USER=docker
      - POSTGRES_PASSWORD=docker
      - POSTGRES_TCP_PORT=5432
```

#### 📖 Explanation

* **Networks**

    * `web`: for frontend + Traefik routing.
    * `intercom`: internal network for backend ↔ database communication.

* **Volumes**

    * `order_pg_vol`: persistent storage for PostgreSQL data.

* **Traefik**

    * Acts as a reverse proxy.
    * `--providers.docker.exposedbydefault=false`: only route explicitly labeled containers.
    * Exposes HTTP on port `80` and dashboard on `8080`.

* **orderservice**

    * Built from the Dockerfile.
    * Depends on PostgreSQL being healthy before starting.
    * Exposes API via `orders.localhost` through Traefik.

* **sws (frontend)**

    * Serves static files from `./frontend` on port `80`.
    * Accessible via `localhost` through Traefik.

* **PostgreSQL**

    * Uses `postgres:latest`.
    * Health check ensures readiness before backend connects.
    * Persists data to `order_pg_vol`.

---

## 🧾 Troubleshooting Log
## 🧱 Issue 1 — Build Script Line Endings

### ❌ Error
Initially, Docker build failed with:

```
/bin/bash^M: bad interpreter: No such file or directory
```

### ✅ Fix
Converted all line endings in `scripts/build-application.sh` from **CRLF → LF**.  
(Windows tends to use CRLF, which breaks shell scripts in Linux containers.)

---

## 🧱 Issue 2 — Missing Go Modules

### ❌ Error
After fixing line endings, the build failed with:

```
missing go.sum entry for module providing package github.com/go-chi/chi/v5 ...
```

### ✅ Fix
Modified the build script to ensure Go dependencies were fetched before building.

```bash
#!/bin/sh
set -e
cd /app || exit 1

go mod tidy
go mod download

CGO_ENABLED=0 GOOS=linux go build -o /app/ordersystem
```

Rebuilt the image successfully.

---

## 🧱 Issue 3 — Database Connection Error

### ❌ Error During Startup
Logs showed:

```
failed to connect to `user=docker database=order`:
hostname resolving error: lookup order-postgres on 127.0.0.11:53: server misbehaving
```

### ✅ Root Cause
- Postgres container was **not yet healthy** when `orderservice` started.
- Docker Compose was trying to start the backend **before** the database was ready.

### ✅ Fix
Added a **healthcheck** to the `order-postgres` service and a **dependency condition** to the backend:

```yaml
order-postgres:
  image: postgres:latest
  restart: always
  networks:
    - intercom
  volumes:
    - order_pg_vol:/var/lib/postgresql/data
  ports:
    - "5432:5432"
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U docker -d order"]
    interval: 5s
    timeout: 5s
    retries: 10
  environment:
    - POSTGRES_DB=order
    - POSTGRES_USER=docker
    - POSTGRES_PASSWORD=docker
    - POSTGRES_TCP_PORT=5432
```

```yaml
orderservice:
  depends_on:
    order-postgres:
      condition: service_healthy
```

---

## 🧱 Issue 4 — PostgreSQL 18 Directory Structure Change

### ❌ Error
After upgrading to Postgres 18, the container failed with:

```
Error: in 18+, these Docker images are configured to store database data in a format which is compatible with "pg_ctlcluster"...
```

### ✅ Root Cause
PostgreSQL 18 changed its data directory layout:
- Expected mount at `/var/lib/postgresql`
- Existing setup mounted `/var/lib/postgresql/data`

### ✅ Fix
Removed old volumes and allowed Docker to recreate them:

```bash
docker compose down
docker volume rm skeleton_order_pg_vol
docker compose up -d
```

After this, PostgreSQL initialized successfully with:

```
database system is ready to accept connections
```

---

## 🧱 Issue 5 — Gateway Timeout on Frontend (Traefik)

### ❌ Error
When visiting `http://localhost`, Traefik returned:

```
Gateway Timeout
```

### ✅ Root Cause
Traefik could not route traffic to backend services because:
- Services weren’t properly attached to the `web` network
- Labels were missing or mismatched


```
## ✅ Final Working State

After all fixes:
```bash
docker compose ps
```

Output:

```
NAME                        STATUS             PORTS
orderservice                Up                 3000/tcp
skeleton-order-postgres-1   Up (healthy)       0.0.0.0:5432->5432/tcp
sws                         Up                 80/tcp
traefik                     Up                 0.0.0.0:80->80/tcp, 0.0.0.0:8080->8080/tcp
```

---

## 🌐 Final Working URLs

| Service                  | URL                                                                                                                            | Description                  |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------ | ---------------------------- |
| 🧠 **Backend (OpenAPI)** | [http://orders.localhost/openapi/index.html#/Menu/get_api_menu](http://orders.localhost/openapi/index.html#/Menu/get_api_menu) | Go backend API documentation |
| 🎨 **Frontend (SWS)**    | [http://localhost/](http://localhost/)                                                                                         | Static web UI served by SWS  |
| ⚙️ **Traefik Dashboard** | [http://localhost:8080](http://localhost:8080)                                                                                 | Reverse proxy dashboard      |

---

## 🧾 Summary of Fixes

| Issue | Error Message                  | Root Cause               | Solution                                  |
| ----- | ------------------------------ | ------------------------ | ----------------------------------------- |
| 1     | `bash^M: bad interpreter`      | Wrong line endings       | Converted to LF                           |
| 2     | `missing go.sum entry`         | Missing Go dependencies  | Added `go mod tidy` and `go mod download` |
| 3     | `hostname resolving error`     | DB not ready             | Added healthcheck and `depends_on`        |
| 4     | `PostgreSQL 18 data directory` | Changed folder structure | Removed old volumes                       |
| 5     | `Gateway Timeout`              | Traefik misconfiguration | Fixed labels and networks                 |

---

