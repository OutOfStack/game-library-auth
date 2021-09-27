package database

import (
	"io/ioutil"

	"github.com/jmoiron/sqlx"
)

const path = "./scripts/seed.sql"

// Seed seeds database
func Seed(db *sqlx.DB) error {
	q, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(string(q)); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
