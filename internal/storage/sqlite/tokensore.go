package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/salivare/auth-server/internal/storage"
	"github.com/salivare/auth-server/pkg/model"
)

var _ storage.RefreshTokenRepository = (*TokenStore)(nil)

type TokenStore struct {
	db *sql.DB
}

func NewTokenStore(db *sql.DB) *TokenStore {
	return &TokenStore{db: db}
}

func (ts TokenStore) Save(ctx context.Context, token model.RefreshToken) error {
	_, err := ts.db.ExecContext(
		ctx,
		"INSERT INTO refresh_token(token_hash, user_id, expires_at, create_at) VALUES (?,?,?,?)",
		token.TokenHash,
		token.UserID,
		token.ExpiresAt,
		token.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (ts TokenStore) FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var t model.RefreshToken

	row := ts.db.QueryRowContext(
		ctx,
		`SELECT token_hash, user_id,expires_at, create_at FROM refresh_token WHERE token_hash = ? LIMIT 1`,
		tokenHash,
	)

	if err := row.Scan(&t.TokenHash, &t.UserID, &t.ExpiresAt, &t.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.RefreshToken{}, ErrNotFound
		}
		return &model.RefreshToken{}, fmt.Errorf("query no token hash: %w", err)
	}
	return &t, nil
}

func (ts TokenStore) DeleteByHash(ctx context.Context, tokenHash string) error {
	_, err := ts.db.ExecContext(
		ctx,
		"DELETE FROM refresh_token WHERE token_hash = ?",
		tokenHash,
	)

	if err != nil {
		return err
	}

	return nil
}
