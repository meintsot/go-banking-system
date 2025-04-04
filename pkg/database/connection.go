package database

import (
	"database/sql"
	"fmt"
	"os"
	// Removing the modernc.org/sqlite import to avoid driver conflicts
	// We'll use the GORM-compatible driver only
)

// DBConnection represents a database connection to SQLite
type DBConnection struct {
	db   *sql.DB
	path string
}

// NewDBConnection creates a new SQLite database connection
func NewDBConnection(path string) (*DBConnection, error) {
	// Check if database file exists, if not create it
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("error creating database file: %v", err)
		}
		file.Close()
	}

	// Note: This function is only kept for backwards compatibility
	// We now use the GORM connection instead
	return nil, fmt.Errorf("legacy DB connection no longer supported, use GormDBConnection instead")
}

// GetDB returns the underlying database connection
func (c *DBConnection) GetDB() *sql.DB {
	return c.db
}

// Close closes the database connection
func (c *DBConnection) Close() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

// InitSchema initializes the database schema
func (c *DBConnection) InitSchema() error {
	// This method is kept for backward compatibility
	// We now use GORM auto-migration instead
	return fmt.Errorf("legacy schema initialization no longer supported, use GORM auto-migration instead")
}
