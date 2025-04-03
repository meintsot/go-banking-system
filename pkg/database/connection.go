package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // Pure Go SQLite implementation
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

	// Open database connection with the modernc.org/sqlite driver
	// Uses "sqlite" instead of "sqlite3" as the driver name
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Check connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return &DBConnection{
		db:   db,
		path: path,
	}, nil
}

// GetDB returns the underlying database connection
func (c *DBConnection) GetDB() *sql.DB {
	return c.db
}

// Close closes the database connection
func (c *DBConnection) Close() error {
	return c.db.Close()
}

// InitSchema initializes the database schema
func (c *DBConnection) InitSchema() error {
	// Create customers table
	_, err := c.db.Exec(`
		CREATE TABLE IF NOT EXISTS customers (
			id TEXT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			phone TEXT,
			address TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating customers table: %v", err)
	}

	// Create accounts table
	_, err = c.db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			balance REAL NOT NULL,
			account_type TEXT NOT NULL,
			created_at TEXT NOT NULL,
			last_activity TEXT NOT NULL,
			FOREIGN KEY (customer_id) REFERENCES customers (id)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating accounts table: %v", err)
	}

	// Create transactions table
	_, err = c.db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			amount REAL NOT NULL,
			transaction_type TEXT NOT NULL,
			description TEXT,
			timestamp TEXT NOT NULL,
			destination_account_id TEXT,
			FOREIGN KEY (account_id) REFERENCES accounts (id),
			FOREIGN KEY (destination_account_id) REFERENCES accounts (id)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating transactions table: %v", err)
	}

	// Create indexes for better query performance
	_, err = c.db.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_customer_id ON accounts (customer_id)`)
	if err != nil {
		return fmt.Errorf("error creating index on accounts.customer_id: %v", err)
	}

	_, err = c.db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions (account_id)`)
	if err != nil {
		return fmt.Errorf("error creating index on transactions.account_id: %v", err)
	}

	_, err = c.db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_destination_account_id ON transactions (destination_account_id)`)
	if err != nil {
		return fmt.Errorf("error creating index on transactions.destination_account_id: %v", err)
	}

	return nil
}
