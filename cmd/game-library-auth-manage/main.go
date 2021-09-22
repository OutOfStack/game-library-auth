package main

import (
	"flag"
	"fmt"
	"log"

	cfg "github.com/OutOfStack/game-library-auth/pkg/config"
	"github.com/OutOfStack/game-library-auth/pkg/database"
	migrate "github.com/rubenv/sql-migrate"
)

type config struct {
	DB struct {
		Host       string `mapstructure:"DB_HOST"`
		Name       string `mapstructure:"DB_NAME"`
		User       string `mapstructure:"DB_USER"`
		Password   string `mapstructure:"DB_PASSWORD"`
		RequireSSL bool   `mapstructure:"DB_REQUIRESSL"`
	} `mapstructure:",squash"`
}

func main() {

	config := &config{}
	if err := cfg.LoadConfig(".", "app", "env", config); err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	fmt.Printf("connected to host: %s, database: %s\n", config.DB.Host, config.DB.Name)

	db, err := database.Open(database.Config{
		Host:       config.DB.Host,
		Name:       config.DB.Name,
		User:       config.DB.User,
		Password:   config.DB.Password,
		RequireSSL: config.DB.RequireSSL,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}
	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
		if err != nil {
			log.Fatalf("Error applying migrations: %v. Applied %d migrations", err, n)
		}
		log.Printf("Migration complete. Applied %d migrations", n)
		return
	case "rollback":
		if _, err := migrate.ExecMax(db.DB, "postgres", migrations, migrate.Down, 1); err != nil {
			log.Fatalf("Error rolling back last migration: %v", err)
		}
		log.Print("Migration rollback complete")
		return
	default:
		log.Print("Unknown command")
	}
}
