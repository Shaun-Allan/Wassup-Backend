# Start from the official Go image
FROM golang:1.21 as builder

WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o app .

# Start a minimal final image
FROM debian:bullseye-slim

WORKDIR /root/

COPY --from=builder /app/app .

EXPOSE 5002

CMD ["./app"]
