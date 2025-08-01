package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	migrationsDir = "scripts/migrations"
	dbDialect     = "postgres"
)

func main() {
	dsn := os.Getenv("DB_DSN")

	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	flag.Parse()

	switch flag.Arg(0) {
	case "migrate":
		if dsn == "" {
			log.Fatal("DB_DSN environment variable is required")
		}
		if err := applyMigrations(dsn, migrations); err != nil {
			log.Fatalf("Apply migrations error: %v", err)
		}
	case "rollback":
		if dsn == "" {
			log.Fatal("DB_DSN environment variable is required")
		}
		if err := rollbackMigration(dsn, migrations); err != nil {
			log.Fatalf("Rollback migration error: %v", err)
		}
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

func applyMigrations(dsn string, migrations *migrate.FileMigrationSource) error {
	db := connectDB(dsn)
	defer func() {
		if cErr := db.Close(); cErr != nil {
			log.Printf("can't close database: %v", cErr)
		}
	}()

	n, err := migrate.Exec(db.DB, dbDialect, migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("%v. Applied %d migrations", err, n)
	}
	fmt.Printf("Migration complete. Applied %d migrations\n", n)
	return nil
}

func rollbackMigration(dsn string, migrations *migrate.FileMigrationSource) error {
	db := connectDB(dsn)
	defer func() {
		if cErr := db.Close(); cErr != nil {
			log.Printf("can't close database: %v", cErr)
		}
	}()

	n, err := migrate.ExecMax(db.DB, dbDialect, migrations, migrate.Down, 1)
	if err != nil {
		return err
	}
	if n == 0 {
		fmt.Println("There is no applied migrations to rollback")
	} else {
		fmt.Println("Migration rollback complete")
	}
	return nil
}

func keygen() {
	if err := crypto.KeyGen(); err != nil {
		log.Fatalf("Error creating private/public keypair: %v", err)
	}
	fmt.Println("Private/public key files successfully created")
}
