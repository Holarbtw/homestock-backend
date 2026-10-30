FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY pkg ./pkg
COPY catalog-service ./catalog-service

WORKDIR /app/catalog-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /catalog-service ./cmd

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /catalog-service /app/catalog-service

EXPOSE 9090

CMD ["/app/catalog-service"]
