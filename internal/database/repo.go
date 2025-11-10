package database

import (
	"context"

	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var (
	tracer = otel.Tracer("db")
)

// UserRepo manages API for user access
type UserRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

// NewUserRepo constructs a user Repo
func NewUserRepo(db *sqlx.DB, log *zap.Logger) *UserRepo {
	return &UserRepo{
		db:  db,
		log: log,
	}
}

// RunWithTx runs a function in a transaction
func (r *UserRepo) RunWithTx(ctx context.Context, f func(context.Context) error) error {
	// check if we're already in a transaction
	if _, ok := database.TxFromContext(ctx); ok {
		// just exec func
		return f(ctx)
	}

	// begin new tx
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := database.WithTx(ctx, tx)
	err = f(txCtx)
	if err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			r.log.Error("rolling back tx", zap.Error(txErr))
			return err
		}
		return err
	}

	return tx.Commit()
}

// query returns the appropriate executor - transaction if present in context, otherwise the database
func (r *UserRepo) query() database.Querier {
	return database.NewQuerier(r.db)
}
