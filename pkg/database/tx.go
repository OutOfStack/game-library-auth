package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type contextKey string

const txKey contextKey = "tx"

// WithTx adds a transaction to the context
func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFromContext retrieves the transaction from context if present
func TxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(*sqlx.Tx)
	return tx, ok
}

// Querier represents types that can execute SQL queries with context already embedded
type Querier interface {
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Ex wraps db/tx with context
type Ex struct {
	db Executor
}

// NewQuerier creates a new Querier
func NewQuerier(db Executor) Querier {
	return &Ex{
		db: db,
	}
}

// Executor represents database/transaction interface
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Exec executes a query with context
func (e *Ex) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.ExecContext(ctx, query, args...)
	}
	return e.db.ExecContext(ctx, query, args...)
}

// Get retrieves a single row with context
func (e *Ex) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return e.db.GetContext(ctx, dest, query, args...)
}
