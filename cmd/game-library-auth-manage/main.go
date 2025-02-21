package main

import (
	"flag"
	"fmt"
	"log"

	conf "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

const migrationsDir string = "scripts/migrations"

func main() {
	cfg, err := conf.Load()
	if err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		db := connectDB(cfg.DB.DSN)
		defer db.Close()
		applyMigrations(db, migrations)
	case "rollback":
		db := connectDB(cfg.DB.DSN)
		defer db.Close()
		rollbackMigration(db, migrations)
	case "keygen":
		keygen()
	default:
		fmt.Println("Unknown command, available commands:")
		fmt.Println("migrate: applies all migrations to database")
		fmt.Println("rollback: roll backs one last migration of database")
		fmt.Println("keygen: creates private/public key pair files")
	}
}

func connectDB(dsn string) *sqlx.DB {
	db, err := database.New(dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func applyMigrations(db *sqlx.DB, migrations *migrate.FileMigrationSource) {
	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Fatalf("Error applying migrations: %v.\nApplied %d migrations", err, n)
	}
	fmt.Printf("Migration complete. Applied %d migrations\n", n)
}

func rollbackMigration(db *sqlx.DB, migrations *migrate.FileMigrationSource) {
	n, err := migrate.ExecMax(db.DB, "postgres", migrations, migrate.Down, 1)
	if err != nil {
		log.Fatalf("Error rolling back last migration: %v", err)
	}
	if n == 0 {
		fmt.Println("There is no applied migrations to rollback")
	} else {
		fmt.Println("Migration rollback complete")
	}
}

func keygen() {
	if err := crypto.KeyGen(); err != nil {
		log.Fatalf("Error creating private/public keypair: %v", err)
	}
	fmt.Println("Private/public key files successfully created")
}
