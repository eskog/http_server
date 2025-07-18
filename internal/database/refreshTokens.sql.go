// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: refreshTokens.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const dropAllTokens = `-- name: DropAllTokens :exec
TRUNCATE TABLE refresh_tokens CASCADE
`

func (q *Queries) DropAllTokens(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, dropAllTokens)
	return err
}

const getOneRefreshToken = `-- name: GetOneRefreshToken :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at from refresh_tokens
WHERE token = $1
`

func (q *Queries) GetOneRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getOneRefreshToken, token)
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

const insertRefreshToken = `-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at
`

type InsertRefreshTokenParams struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, insertRefreshToken, arg.Token, arg.UserID, arg.ExpiresAt)
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

const revokeToken = `-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1
`

func (q *Queries) RevokeToken(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, revokeToken, token)
	return err
}
