FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o shop-service ./cmd/service/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/shop-service .
COPY .env .env

EXPOSE 8080

CMD ["./shop-service"]