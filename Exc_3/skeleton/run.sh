#!/bin/sh

# todo
# docker build
set -e

# Removing the old DB container & volume
docker rm -f orders-db || true
docker volume rm pgdata || true

# create new volume
docker volume create pgdata

# docker run db
# Start PostgreSQL container
docker run -d \
  --name orders-db \
  -e POSTGRES_DB=order \
  -e POSTGRES_USER=docker \
  -e POSTGRES_PASSWORD=docker \
  -e PGDATA=/var/lib/postgresql/18/docker \
  -v pgdata:/var/lib/postgresql/18/docker \
  -p 5432:5432 \
  postgres:18

# Wait a few seconds for DB to initialize
sleep 5



# docker run orderservice

# Build Orderservice Docker image
docker build -f ../Dockerfile -t orderservice:latest ..


# Remove old Orderservice container & run new one
docker rm -f orderservice || true

docker run -d \
  --name orderservice \
  --link orders-db:orders-db \
  -e POSTGRES_USER=docker \
  -e POSTGRES_PASSWORD=docker \
  -e POSTGRES_DB=order \
  -e POSTGRES_TCP_PORT=5432 \
  -e DB_HOST=orders-db \
  -p 3000:3000 \
  orderservice:latest

# show if both containers started
docker ps

# show Orderservice logs
docker logs -f orderservice



# to savely exit the Docker Containers use: docker stop container_name
# to savely exit the Docker Containers use: docker stop container_name
# start them again:
## docker start orderservice
## docker start orders-db

# in web look at http://localhost:3000/