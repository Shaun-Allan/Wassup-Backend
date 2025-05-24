# Use an official Golang image as the base
FROM golang:1.21 as builder

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
RUN go build -o main .

# Final image
FROM debian:bookworm-slim

# Set working directory
WORKDIR /root/

# Copy the built binary from the builder
COPY --from=builder /app/main .

# Expose the application port
EXPOSE 5004

# Command to run the application
CMD ["./main"]
