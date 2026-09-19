package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB initializes the database connection
func InitDB(dsn string) (*gorm.DB, error) {
	if os.Getenv("USE_SQLITE") == "true" {
		log.Println("Using SQLite database for local development")
		db, err := gorm.Open(sqlite.Open("barcvvr.db"), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to local sqlite database: %w", err)
		}
		return db, nil
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connection established")
	return db, nil
}
