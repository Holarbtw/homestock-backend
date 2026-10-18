package postgres

import (
	"context"
	"time"

	"github.com/Holarbtw/homestock-backend/auth-service/internal/domain"
	"github.com/Holarbtw/homestock-backend/pkg/apperrors"
	pkgPostgres "github.com/Holarbtw/homestock-backend/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var authConstraintMap = map[string]*apperrors.CustomError{
	"idx_users_unique_active_email": domain.ErrEmailAlreadyExists,
	"refresh_tokens_token_hash_key": domain.ErrSessionRevoked,
}

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) getExecutor(ctx context.Context) pkgPostgres.QueryExecutor {
	return pkgPostgres.GetExecutor(ctx, r.pool)
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, nickname, timezone string) (*domain.User, error) {
	query := `
		insert into users (email, password_hash, nickname, timezone, is_guest)
		values ($1, $2, $3, $4, false)
		returning id, email, password_hash, nickname, timezone, is_guest, created_at, updated_at, deleted_at
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, email, passwordHash, nickname, timezone)
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	return toDomainUser(&record), nil
}

func (r *UserRepository) CreateGuest(ctx context.Context, nickname, timezone string) (*domain.User, error) {
	query := `
		insert into users (email, password_hash, nickname, timezone, is_guest)
		values (null, null, $1, $2, true)
		returning id, email, password_hash, nickname, timezone, is_guest, created_at, updated_at, deleted_at
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, nickname, timezone)
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	return toDomainUser(&record), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		select id, email, password_hash, nickname, timezone, is_guest, created_at, updated_at, deleted_at
		from users
		where email = $1 and deleted_at is null
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, email)
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	return toDomainUser(&record), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		select id, email, password_hash, nickname, timezone, is_guest, created_at, updated_at, deleted_at
		from users
		where id = $1 and deleted_at is null
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, id)
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	return toDomainUser(&record), nil
}

func (r *UserRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		select id, email, password_hash, nickname, timezone, is_guest, created_at, updated_at, deleted_at
		from users
		where id = $1 and deleted_at is null
		for update
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, id)
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	return toDomainUser(&record), nil
}

func (r *UserRepository) LinkGuestAccount(ctx context.Context, userID uuid.UUID, email, passwordHash, nickname string) error {
	query := `
		update users
		set email = $1, password_hash = $2, nickname = $3, is_guest = false, updated_at = now()
		where id = $4 and deleted_at is null
	`

	tag, err := r.getExecutor(ctx).Exec(ctx, query, email, passwordHash, nickname, userID)
	if err != nil {
		return apperrors.MapPgError(err, authConstraintMap)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *UserRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*domain.RefreshToken, error) {
	query := `
		insert into refresh_tokens (user_id, token_hash, expires_at)
		values ($1, $2, $3)
		returning id, user_id, token_hash, expires_at, created_at, revoked_at
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[refreshTokenDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	return toDomainRefreshToken(&record), nil
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		select id, user_id, token_hash, expires_at, created_at, revoked_at
		from refresh_tokens
		where token_hash = $1
	`

	rows, err := r.getExecutor(ctx).Query(ctx, query, tokenHash)
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[refreshTokenDB])
	if err != nil {
		return nil, apperrors.MapPgError(err, authConstraintMap)
	}

	return toDomainRefreshToken(&record), nil
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
		update refresh_tokens
		set revoked_at = now()
		where token_hash = $1 and revoked_at is null
	`

	_, err := r.getExecutor(ctx).Exec(ctx, query, tokenHash)
	if err != nil {
		return apperrors.MapPgError(err, authConstraintMap)
	}

	return nil
}

func (r *UserRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `
		update refresh_tokens
		set revoked_at = now()
		where user_id = $1 and revoked_at is null
	`

	_, err := r.getExecutor(ctx).Exec(ctx, query, userID)
	if err != nil {
		return apperrors.MapPgError(err, authConstraintMap)
	}

	return nil
}
