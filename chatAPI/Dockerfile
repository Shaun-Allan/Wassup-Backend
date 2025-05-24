# Build stage
FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o chatserver .

# Final image
FROM gcr.io/distroless/base-debian11

WORKDIR /app
COPY --from=builder /app/chatserver .

EXPOSE 5003

CMD ["/app/chatserver"]
