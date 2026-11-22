FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY pkg ./pkg
COPY inventory-service ./inventory-service

WORKDIR /app/inventory-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /inventory-service ./cmd

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /inventory-service /app/inventory-service

EXPOSE 8081

CMD ["/app/inventory-service"]
