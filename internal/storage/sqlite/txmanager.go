package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLTxManager struct {
	db *sql.DB
}

// NewTxManager creates a new transaction manager instance.
func NewTxManager(db *sql.DB) *SQLTxManager {
	return &SQLTxManager{db: db}
}

type contextKey string

const txKey contextKey = "tx"

// RunInTx initiates transaction
func (m *SQLTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sqlite: failed to begin transaction: %w", err)
	}

	// Put transaction in new context
	ctxWithTx := context.WithValue(ctx, txKey, tx)

	err = fn(ctxWithTx)

	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sqlite: transaction commit failed: %w", err)
	}

	return nil
}
