# Stage 1: Build
FROM golang:1.21 as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go application
RUN go build -o group-server

# Stage 2: Run
FROM debian:bullseye-slim

WORKDIR /app

# Install CA certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from builder
COPY --from=builder /app/group-server .

# Expose the application port
EXPOSE 5005

# Command to run the application
CMD ["./group-server"]
