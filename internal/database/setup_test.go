package database_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

const (
	migrationsDir = "../../scripts/migrations"
	pg            = "postgres"
)

var db *sqlx.DB
var dsn string

func TestMain(m *testing.M) {
	ctx := context.Background()

	// run postgres docker container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("auth_test"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(1*time.Minute)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}
	log.Println("Repo tests: Docker container started")

	dsn, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	// connect to database
	db, err = sqlx.Connect(pg, dsn)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}
	log.Println("Repo tests: Database connection established")

	// run tests in current package
	code := m.Run()

	// can't defer because of os.Exit
	db.Close()

	if err = testcontainers.TerminateContainer(pgContainer); err != nil {
		log.Printf("failed to terminate postgres container: %v", err)
	}
	log.Println("Repo tests: Docker container deleted")

	os.Exit(code)
}

// setup applies all migrations to the test database
func setup(t *testing.T) *database.UserRepo {
	t.Helper()

	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	n, err := migrate.Exec(db.DB, pg, migrations, migrate.Up)
	if err != nil {
		t.Fatalf("failed to apply migrations: %v. Applied %d migrations", err, n)
	}

	return database.NewUserRepo(db, zap.NewNop())
}

// teardown rolls back all migrations from the test database
func teardown(t *testing.T) {
	t.Helper()

	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	// get current migration count
	records, err := migrate.GetMigrationRecords(db.DB, pg)
	if err != nil {
		t.Fatalf("failed to get migration records: %v", err)
	}

	// roll back all migrations
	n, err := migrate.ExecMax(db.DB, pg, migrations, migrate.Down, len(records))
	if err != nil {
		t.Fatalf("failed to rollback migrations: %v. Rolled back %d migrations", err, n)
	}
}
