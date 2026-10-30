FROM golang:1.26-alpine AS builder

RUN CGO_ENABLED=0 go install \
    -tags='no_mysql no_sqlite3 no_ydb no_clickhouse no_mssql no_vertica no_libsql' \
    github.com/pressly/goose/v3/cmd/goose@v3.26.0

FROM alpine:3.23

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY catalog-service/migrations /migrations

USER 65532:65532
ENV GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=/migrations
ENTRYPOINT ["goose", "-env=none"]
CMD ["up"]
