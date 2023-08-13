package database

import (
	"github.com/OutOfStack/game-library-auth/scripts"
	"github.com/jmoiron/sqlx"
)

// Seed seeds database
func Seed(db *sqlx.DB) error {
	q := scripts.SeedSQL

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(q); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return rErr
		}
		return err
	}

	return tx.Commit()
}
