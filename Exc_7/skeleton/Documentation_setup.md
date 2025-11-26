Here’s your Docker Swarm Deployment Documentation formatted as a clean **Markdown (.md)** file. You can click/open the card above to download it.


# Docker Swarm Deployment Documentation

## 1. Introduction

This document describes the complete setup and deployment of the **SBD AIS Exercise** application using Docker Swarm.  
The stack consists of:

- Traefik (reverse proxy / ingress controller)  
- Frontend  
- OrderService  
- Postgres database  
- MinIO (S3 storage)  

All workloads run on a single-node Docker Swarm cluster.

---

## 2. Initialize Docker Swarm

Start Docker Desktop and run:

```bash
docker swarm init
```

**Output:**
```
Swarm initialized: current node (z92adx8ysxnj678jye9uuw082) is now a manager.
```

Verify that Swarm mode is active:

```bash
docker info | grep Swarm
```

**Result:**
```
Swarm: active
```

List nodes:

```bash
docker node ls
```

**Output:**
```
ID                            HOSTNAME         STATUS    AVAILABILITY   MANAGER STATUS
z92adx8ysxnj678jye9uuw082 *   docker-desktop   Ready     Active         Leader
```

The cluster consists of 1 manager node (Docker Desktop).

---

## 3. Docker Secrets

The stack uses Docker Secrets for:

- Postgres username
- Postgres password
- S3 access key
- S3 secret key

Secrets are stored in:

- `docker/postgres_user_secret`
- `docker/postgres_password_secret`
- `docker/s3_user_secret`
- `docker/s3_password_secret`

These are automatically loaded during deployment.

---

## 4. Docker Networks

Two overlay networks are created automatically:

| Network   | Purpose                                |
|-----------|----------------------------------------|
| web       | Public HTTP traffic routed by Traefik  |
| intercom  | Internal service communication         |

Since Swarm is single-node, overlay networks still work.

---

## 5. docker-compose.yml

Below is the full stack configuration used for Docker Swarm:

```yaml
version: "3.9"  # Specifies the Docker Compose file format version

networks:  # Define custom networks for inter-service communication
  web:  # Network for services exposed to the outside (frontend, Traefik)
    driver: overlay  # Overlay network allows cross-node communication in swarm
    attachable: true  # Allow standalone containers to connect
  intercom:  # Network for internal service communication (backend services)
    driver: overlay
    attachable: true

volumes:  # Define persistent volumes for data storage
  order_pg_vol:  # Volume for PostgreSQL database
  minio_vol:  # Volume for MinIO object storage

secrets:  # Define Docker secrets for sensitive information
  postgres_user:  # Secret for PostgreSQL username
    file: docker/postgres_user_secret  # File containing the username
  postgres_password:  # Secret for PostgreSQL password
    file: docker/postgres_password_secret
  s3_user:  # Secret for MinIO/S3 access key
    file: docker/s3_user_secret
  s3_password:  # Secret for MinIO/S3 secret key
    file: docker/s3_password_secret

services:  # Define all services in the stack
  traefik:  # Traefik reverse proxy service
    image: traefik:v3.6.1  # Traefik Docker image
    command:  # Additional Traefik CLI options
      - --api.insecure=false  # Disable insecure API
      - --api.dashboard=true  # Enable Traefik dashboard
      - --providers.swarm=true  # Use Docker Swarm provider
      - --providers.docker.exposedbydefault=false  # Only expose services with labels
      - --entrypoints.web.address=:80  # Define HTTP entrypoint on port 80
      - --log.level=INFO  # Set log level to INFO
    ports:  # Expose ports on the host
      - target: 80  # Internal port
        published: 80  # Host port
        protocol: tcp  # TCP protocol
        mode: ingress  # Ingress mode for swarm routing
      - target: 8080  # Internal dashboard port
        published: 8080  # Host port
        protocol: tcp
        mode: ingress
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro  # Allow Traefik to access Docker API
    networks:
      - web
      - intercom
    deploy:  # Swarm deployment configuration
      replicas: 1  # Only one Traefik instance
      placement:
        constraints:
          - node.role == manager  # Run only on manager nodes
      labels:  # Traefik labels to configure routing
        - traefik.enable=true  # Enable Traefik for this service
        - traefik.http.routers.traefik-dashboard.rule=Host(`localhost`) && PathPrefix(`/dashboard`)  # URL rule
        - traefik.http.routers.traefik-dashboard.entrypoints=web  # Use web entrypoint
        - traefik.http.services.traefik-dashboard.loadbalancer.server.port=80  # Target port

  frontend:  # Frontend web application
    image: ghcr.io/ddibiasi/sbd-ais-exercise/frontend:1.0
    networks:
      - web
    deploy:
      mode: global  # Deploy one instance per node
      labels:
        - traefik.enable=true
        - traefik.http.routers.frontend.rule=Host(`localhost`)  # Route requests to localhost
        - traefik.http.routers.frontend.entrypoints=web
        - traefik.http.services.frontend.loadbalancer.server.port=80
      restart_policy:
        condition: on-failure  # Restart only on failure

  orderservice:  # Backend order processing service
    image: ghcr.io/ddibiasi/sbd-ais-exercise/orderservice:1.0
    command: [ "/app/ordersystem" ]  # Command to start the service
    networks:
      - web
      - intercom
    secrets:  # Attach secrets to the container
      - postgres_user
      - postgres_password
      - s3_user
      - s3_password
    environment:  # Environment variables for service configuration
      - DB_HOST=postgres  # Database hostname
      - PGPORT=5432  # Database port
      - POSTGRES_DB=order  # Database name
      - POSTGRES_USER_FILE=/run/secrets/postgres_user  # Secret file for user
      - POSTGRES_PASSWORD_FILE=/run/secrets/postgres_password
      - S3_ENDPOINT=minio:8500  # S3/MinIO endpoint
      - S3_ACCESS_KEY_ID_FILE=/run/secrets/s3_user
      - S3_SECRET_ACCESS_KEY_FILE=/run/secrets/s3_password
    deploy:
      mode: global
      labels:
        - traefik.enable=true
        - traefik.http.routers.orders.rule=Host(`orders.localhost`)  # Route requests to orders.localhost
        - traefik.http.routers.orders.entrypoints=web
        - traefik.http.services.orders.loadbalancer.server.port=3000  # Internal service port
      restart_policy:
        condition: any
        delay: 5s
        max_attempts: 10
        window: 60s

  postgres:  # PostgreSQL database service
    image: postgres:18
    volumes:
      - order_pg_vol:/var/lib/postgresql  # Mount persistent volume
    networks:
      - intercom
    secrets:
      - postgres_user
      - postgres_password
    environment:
      - POSTGRES_DB=order
      - POSTGRES_USER_FILE=/run/secrets/postgres_user
      - POSTGRES_PASSWORD_FILE=/run/secrets/postgres_password
      - PGPORT=5432
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager  # Run only on manager nodes
      restart_policy:
        condition: on-failure

  minio:  # MinIO object storage service
    image: minio/minio:latest
    command: [ "server", "--address", ":8500", "/data" ]  # Start MinIO server
    volumes:
      - minio_vol:/data  # Mount persistent storage
    networks:
      - intercom
    secrets:
      - s3_user
      - s3_password
    environment:
      - MINIO_ROOT_USER_FILE=/run/secrets/s3_user  # Secret for root user
      - MINIO_ROOT_PASSWORD_FILE=/run/secrets/s3_password
    deploy:
      replicas: 1
      placement:
        constraints:
          - node.role == manager
      restart_policy:
        condition: on-failure
```

---

## 6. Deploying the Stack

Run:

```bash
docker stack deploy -c docker-compose.yml sbd
```

**Output:**
```
Creating network sbd_web
Creating network sbd_intercom
Creating secret sbd_postgres_user
Creating secret sbd_postgres_password
Creating secret sbd_s3_user
Creating secret sbd_s3_password
Creating service sbd_traefik
Creating service sbd_frontend
Creating service sbd_orderservice
Creating service sbd_postgres
Creating service sbd_minio
```

---

## 7. Check Running Services

```bash
docker service ls
```

**Example output:**
```
NAME               MODE         REPLICAS   IMAGE
sbd_frontend       global       1/1        frontend:1.0
sbd_minio          replicated   1/1        minio:latest
sbd_orderservice   global       1/1        orderservice:1.0
sbd_postgres       replicated   1/1        postgres:18
sbd_traefik        replicated   1/1        traefik:v3.6.1
```

---

## 8. Logs and Debugging

Example of checking logs:

```bash
docker service logs sbd_orderservice
```

**Observed error during first startup:**
```
S3 is not reachable, timeout after waiting for 10 seconds
```

This happens because `orderservice` starts before MinIO is ready.  
Swarm automatically retries due to restart policies.

After MinIO becomes available, `orderservice` enters **Running** state.

---

## 9. Accessing the Application

| Component          | URL                          |
|--------------------|------------------------------|
| Frontend           | http://localhost             |
| Traefik Dashboard  | http://localhost/dashboard   |
| Orderservice       | http://orders.localhost      |
| MinIO UI           | (not exposed through Traefik)|

---

## 10. Summary

The system is now fully deployed on a single-node Docker Swarm cluster.

- Traefik is handling routing  
- Frontend is reachable through **http://localhost**  
- Orderservice connects to Postgres and MinIO  
- All secrets are managed through Docker Swarm  
- All services run successfully as seen with `docker stack ps sbd`
```
