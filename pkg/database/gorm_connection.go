package database

import (
	"bankingsystem/pkg/models"
	"fmt"
	"os"

	"github.com/glebarez/sqlite" // Pure Go SQLite driver for GORM
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDBConnection represents a database connection using GORM
type GormDBConnection struct {
	db   *gorm.DB
	path string
}

// NewGormDBConnection creates a new GORM database connection
func NewGormDBConnection(path string) (*GormDBConnection, error) {
	// Check if database file exists, if not create it
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("error creating database file: %v", err)
		}
		file.Close()
	}

	// Set up GORM configuration
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Open database connection with GORM and the pure Go SQLite driver
	db, err := gorm.Open(sqlite.Open(path), config)
	if err != nil {
		return nil, fmt.Errorf("error opening database with GORM: %v", err)
	}

	return &GormDBConnection{
		db:   db,
		path: path,
	}, nil
}

// GetDB returns the underlying GORM database connection
func (c *GormDBConnection) GetDB() *gorm.DB {
	return c.db
}

// Close closes the database connection
func (c *GormDBConnection) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("error getting SQL DB: %v", err)
	}
	return sqlDB.Close()
}

// InitSchema initializes the database schema using GORM's auto-migration
func (c *GormDBConnection) InitSchema() error {
	// Automatically migrate the schema
	err := c.db.AutoMigrate(&models.Customer{}, &models.Account{}, &models.Transaction{})
	if err != nil {
		return fmt.Errorf("error migrating schema: %v", err)
	}

	return nil
}
