// Package db manages the lifecycle of the PostgreSQL connection pool and transaction bounds.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"skillbridge-backend/pkg/config"
	"skillbridge-backend/pkg/logger"
	"time"

	_ "github.com/lib/pq"
)

// DB represents the active PostgreSQL connection pool.
var DB *sql.DB

// Init configures connection limits and verifies relational database accessibility.
func Init() {
	cfg := config.AppConfig
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSslMode)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to open database connection", err)
		logPanic(err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)

	err = DB.Ping()
	if err != nil {
		logger.Error("Failed to ping database", err)
		logPanic(err)
	}

	logger.Info("Successfully connected to PostgreSQL database")

	InitSchema()
}

// InitSchema loads and runs SQL scripts to initialize schema tables automatically on boot.
func InitSchema() {
	schemaFile := filepath.Join("migrations", "000001_init_schema.up.sql")
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		logger.Warn("Could not find migration file, skipping setup: " + err.Error())
		return
	}

	_, err = DB.Exec(string(content))
	if err != nil {
		logger.Error("Failed to execute database migrations", err)
		return
	}

	logger.Info("Database migrations initialized successfully")
}

// WithTransaction runs queries inside a database transaction, rolling back automatically on failure or panics.
func WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func logPanic(err error) {
	panic(err)
}
