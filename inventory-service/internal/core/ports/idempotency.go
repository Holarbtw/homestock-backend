package ports

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

type IdempotencyResult struct {
	// Data contains an encoded use-case result, independent of the transport format.
	Data []byte
}

type IdempotencyRecord struct {
	SpaceID     uuid.UUID
	RequestKey  string
	CommandType string
	RequestHash [sha256.Size]byte
	// Result is nil while the reservation has not been finalized.
	Result    *IdempotencyResult
	CreatedAt time.Time
}
