package postgres

import (
	"time"

	"github.com/google/uuid"
)

type userDB struct {
	ID           uuid.UUID  `db:"id"`
	Email        *string    `db:"email"`
	PasswordHash *string    `db:"password_hash"`
	Nickname     string     `db:"nickname"`
	Timezone     string     `db:"timezone"`
	IsGuest      bool       `db:"is_guest"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

type refreshTokenDB struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}
