# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

# Copy go.mod and go.sum first and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -o user-search-app

# Final image
FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/user-search-app .

EXPOSE 5001

CMD ["./user-search-app"]
