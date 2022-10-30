package database

import (
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // register postgres driver
)

// Config stores connection settings
type Config struct {
	Host       string
	Name       string
	User       string
	Password   string
	RequireSSL bool
}

// Open opens connection with database
func Open(cfg Config) (*sqlx.DB, error) {
	query := url.Values{}
	if cfg.RequireSSL {
		query.Set("sslmode", "require")
	} else {
		query.Set("sslmode", "disable")
	}

	query.Set("timezone", "utc")

	conn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: query.Encode(),
	}

	return sqlx.Connect("postgres", conn.String())
}

// StatusCheck returns nil if connection with db is ok
func StatusCheck(db *sqlx.DB) error {
	return db.Ping()
}
