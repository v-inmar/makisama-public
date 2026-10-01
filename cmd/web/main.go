package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"app/internal/constants"
	"app/internal/database"
	"app/internal/server"
	logger_utils "app/internal/utilities/logger"
	template_utils "app/internal/utilities/templates"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {

	// load environment variables
	// silent load so won't fail when using docker
	_ = godotenv.Load()
	// if err := godotenv.Load(); err != nil {
	// 	// log.Println("No .env file found")
	// 	panic("Unable to load env variables")
	// }

	// check it loads env var when using docker
	if os.Getenv("DB_URL") == "" {
		panic("Required environment variable DB_URL is not set")
	}

	env := os.Getenv("ENVIRONMENT")

	if err := createLogDirectory(); err != nil {
		panic("Unable to create log directory")
	}

	appLogger, logCloser := logger_utils.NewZapLogger(constants.LogDirectory, "app", os.Getenv("ENVIRONMENT"))
	defer appLogger.Sync() // Flushes buffered log lines to disk before exit
	defer logCloser.Close()

	if env != "production" {
		log.Println("Application initialized successfully. Rolling logs active!")
	}

	// grab db conn string
	dsn := os.Getenv("DB_URL")
	db, err := database.ConnectPostgres(dsn)
	if err != nil {
		log.Fatalf("Critical failure initializing database: %v", err)
	}
	defer db.Close() // Ensure the entire pool closes gracefully when the app stops

	// Create session table
	if err := createSessionTableInDb(db); err != nil {
		log.Panicf("failed to create session table in the database. %v", err)
	}

	// Create session manager
	session := createAndConfigureSessionManager(db)
	if env != "production" {
		log.Println("Session Manager created and configured")
	}

	// Run migration
	if err := runMigration(dsn); err != nil {
		log.Panicf("failed to run migration. %v", err)
	}
	if env != "production" {
		log.Println("Database migrations applied successfully")
	}

	// create image board directory
	if err := createBoardImageDirectory(); err != nil {
		log.Panicf("failed to create board image directory. %v", err)
	}

	// assemble the Infrastructure Configuration Package
	srvConfig := &server.Config{
		DB:        db,
		Session:   session,
		Templates: template_utils.ParseTemplates(),
		Logger:    appLogger,
	}

	// pass configuration to the router shell
	appRouter := server.ConfigureRoutes(srvConfig)
	host := os.Getenv("HOST")
	if host == "" {
		panic("Unable to get host from env variable")
	}

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		panic("Unable to get port from env variables")
	}

	addr := fmt.Sprintf("%s%s", host, port)

	if err := createSessionTableInDb(db); err != nil {
		log.Printf("Server executing cleanly on %s\n", addr)
	}

	// fire up the actual HTTP Network Server Engine
	log.Println("Server started...")
	if err := http.ListenAndServe(addr, appRouter); err != nil {
		panic(err)
	}

}

func createAndConfigureSessionManager(db *sqlx.DB) *scs.SessionManager {
	session := scs.New()

	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = false
	session.Cookie.Name = constants.SessionName
	session.Store = postgresstore.New(db.DB)
	session.Cookie.HttpOnly = true

	if os.Getenv("ENVIRONMENT") == "production" {
		session.Cookie.SameSite = http.SameSiteLaxMode
		session.Cookie.Secure = true
	} else {
		session.Cookie.SameSite = http.SameSiteLaxMode
		session.Cookie.Secure = false
	}

	// session := scs.New()
	// session.Lifetime = 24 * time.Hour
	// if os.Getenv("ENVIRONMENT") == "production"{

	// }

	// session.Cookie.Persist = false // set persistence in handler
	// session.Cookie.SameSite = http.SameSiteDefaultMode
	// session.Cookie.Secure = false            // TODO: change to true on prod
	// session.Store = postgresstore.New(db.DB) // using postgres as store
	// session.Cookie.Name = constants.SessionName

	// session.Cookie.Name = constants.SessionName
	// session.Cookie.Secure = true
	// session.Cookie.HttpOnly = true
	// session.Cookie.SameSite = http.SameSiteLaxMode

	// session.Lifetime = 24 * time.Hour
	// session.Store = postgresstore.New(db.DB)

	// if os.Getenv("ENVIRONMENT") == "production"{
	// 	session := scs.New()
	// 	session.Lifetime = 24 * time.Hour
	// 	session.Cookie.Persist = false // set persistence in handler
	// 	session.Cookie.SameSite = http.SameSiteDefaultMode
	// 	session.Cookie.Secure = false            // TODO: change to true on prod
	// 	session.Store = postgresstore.New(db.DB) // using postgres as store
	// 	session.Cookie.Name = constants.SessionName
	// }

	return session
}

func createSessionTableInDb(db *sqlx.DB) error {
	// session table create query
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		data BYTEA NOT NULL,
		expiry TIMESTAMPTZ NOT NULL
	);
	CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions (expiry);
	`
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil

}

func runMigration(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil

}

func createBoardImageDirectory() error {
	dirPath := constants.BoardUploadImageDirectory
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	return nil
}

func createLogDirectory() error {
	dirPath := constants.LogDirectory
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	return nil
}
