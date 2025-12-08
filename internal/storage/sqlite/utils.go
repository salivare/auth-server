package sqlite

import (
	"context"
	"database/sql"
)

// DBConnector оставляем как есть
type DBConnector interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Storage struct
type Storage struct {
	db *sql.DB
	SQLTxManager
	users  UserRepository
	tokens TokenRepository
}

// NewStorage constructor — возвращаем конкретный тип
func NewStorage(db *sql.DB) *Storage {
	s := &Storage{
		db:           db,
		SQLTxManager: SQLTxManager{db: db},
	}

	s.users = UserRepository{s: s}
	s.tokens = TokenRepository{s: s}

	return s
}

func (s *Storage) Users() *UserRepository {
	return &s.users
}

func (s *Storage) Tokens() *TokenRepository {
	return &s.tokens
}

func (s *Storage) connector(ctx context.Context) DBConnector {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return s.db
}
