package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/salivare/auth-server/internal/storage"
	"github.com/salivare/auth-server/pkg/model"
)

type TokenRepository struct {
	s *Storage
}

func NewTokenRepository(s *Storage) *TokenRepository {
	return &TokenRepository{s: s}
}

func (ts *TokenRepository) Save(ctx context.Context, token model.RefreshToken) error {
	conn := ts.s.connector(ctx)
	_, err := conn.ExecContext(
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

func (ts *TokenRepository) FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	conn := ts.s.connector(ctx)

	var t model.RefreshToken

	row := conn.QueryRowContext(
		ctx,
		`SELECT token_hash, user_id,expires_at, create_at FROM refresh_token WHERE token_hash = ? LIMIT 1`,
		tokenHash,
	)

	if err := row.Scan(&t.TokenHash, &t.UserID, &t.ExpiresAt, &t.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &model.RefreshToken{}, storage.ErrTokenNotFound
		}
		return &model.RefreshToken{}, fmt.Errorf("query no token hash: %w", err)
	}
	return &t, nil
}

func (ts *TokenRepository) DeleteByHash(ctx context.Context, tokenHash string) error {
	conn := ts.s.connector(ctx)

	_, err := conn.ExecContext(
		ctx,
		"DELETE FROM refresh_token WHERE token_hash = ?",
		tokenHash,
	)

	if err != nil {
		return err
	}

	return nil
}
