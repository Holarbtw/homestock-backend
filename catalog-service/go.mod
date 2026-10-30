module github.com/Holarbtw/homestock-backend/catalog-service

go 1.26.4

replace github.com/Holarbtw/homestock-backend/pkg => ../pkg

require (
	github.com/Holarbtw/homestock-backend/pkg v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.11.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)
