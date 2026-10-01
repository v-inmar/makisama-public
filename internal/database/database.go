package database

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Pure Go Postgres driver
)

// ConnectPostgres establishes a secure, verified connection pool with the database
func ConnectPostgres(dsn string) (*sqlx.DB, error) {
	// sqlx.Connect combines opening a connection and pinging it to make sure it works
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool optimization parameters
	db.SetMaxOpenConns(25)                 // Max active connections allowed at once
	db.SetMaxIdleConns(25)                 // Max idle connections kept alive in reserve
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum age of a reusable connection

	return db, nil
}
