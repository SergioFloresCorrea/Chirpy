// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: refresh_tokens.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at)
VALUES(
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at
`

type CreateRefreshTokenParams struct {
	Token     string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	ExpiresAt time.Time
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, createRefreshToken,
		arg.Token,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.ExpiresAt,
	)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const getRefreshTokenByToken = `-- name: GetRefreshTokenByToken :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at FROM refresh_tokens
WHERE token=$1
LIMIT 1
`

func (q *Queries) GetRefreshTokenByToken(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getRefreshTokenByToken, token)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const setRevokeAt = `-- name: SetRevokeAt :exec
UPDATE refresh_tokens
SET revoked_at = $1, updated_at = $2
WHERE token = $3
`

type SetRevokeAtParams struct {
	RevokedAt sql.NullTime
	UpdatedAt time.Time
	Token     string
}

func (q *Queries) SetRevokeAt(ctx context.Context, arg SetRevokeAtParams) error {
	_, err := q.db.ExecContext(ctx, setRevokeAt, arg.RevokedAt, arg.UpdatedAt, arg.Token)
	return err
}
