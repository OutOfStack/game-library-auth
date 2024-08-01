package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	conf "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

const migrationsDir string = "scripts/migrations"

func main() {
	cfg, err := conf.Init()
	if err != nil {
		log.Fatalf("can't parse config: %v", err)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		db := connectDB(cfg.DB)
		defer db.Close()
		applyMigrations(db, migrations)
	case "rollback":
		db := connectDB(cfg.DB)
		defer db.Close()
		rollbackMigration(db, migrations)
	case "seed":
		db := connectDB(cfg.DB)
		defer db.Close()
		seed(db)
	case "keygen":
		keygen()
	default:
		fmt.Println("Unknown command, available commands:")
		fmt.Println("migrate: applies all migrations to database")
		fmt.Println("rollback: roll backs one last migration of database")
		fmt.Println("seed: applies seed data (roles, admin user) to database")
		fmt.Println("keygen: creates private/public key pair files")
	}
}

func connectDB(conf appconf.DB) *sqlx.DB {
	db, err := database.Open(database.Config{
		Host:       conf.Host,
		Name:       conf.Name,
		User:       conf.User,
		Password:   conf.Password,
		RequireSSL: conf.RequireSSL,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("connected to host: %s, database: %s\n", conf.Host, conf.Name)

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

func seed(db *sqlx.DB) {
	if err := database.Seed(db); err != nil {
		log.Fatalf("Error applying seeds: %v", err)
	}
	log.Println("Seed data successfully inserted")
}

func keygen() {
	if err := crypto.KeyGen(); err != nil {
		log.Fatalf("Error creating private/public keypair: %v", err)
	}
	fmt.Println("Private/public key files successfully created")
}
