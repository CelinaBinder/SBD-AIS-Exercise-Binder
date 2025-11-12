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