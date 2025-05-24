# Use Go official image for building
FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app .

# Final image
FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 5000

CMD ["/app/app"]
