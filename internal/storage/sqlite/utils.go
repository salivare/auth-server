package sqlite

import (
	"context"
	"database/sql"
	"github.com/salivare/auth-server/internal/storage"
)

type DBConnector interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Storage struct
type Storage struct {
	db *sql.DB
	sqlTxManager
	users  userRepository
	tokens tokenRepository
}

// NewStorage constructor
func NewStorage(db *sql.DB) storage.Storage {
	s := &Storage{
		db:           db,
		sqlTxManager: sqlTxManager{db: db},
	}

	s.users = userRepository{s: s}
	s.tokens = tokenRepository{s: s}

	return s
}

func (s *Storage) Users() storage.UserRepository {
	return &s.users
}

func (s *Storage) Tokens() storage.RefreshTokenRepository {
	return &s.tokens
}

func (s *Storage) connector(ctx context.Context) DBConnector {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return s.db
}
