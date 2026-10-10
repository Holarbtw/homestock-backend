FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/main.go

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service

EXPOSE 8080

CMD ["/app/auth-service"]
