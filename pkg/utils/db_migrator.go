package utils

import (
	"bankingsystem/pkg/models"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver with underscore
	"gorm.io/gorm"
)

// DBMigrator is a utility for migrating data from old SQLite database to GORM database
type DBMigrator struct {
	sourceDB *sql.DB
	targetDB *gorm.DB
	logger   *log.Logger
}

// NewDBMigrator creates a new database migrator
func NewDBMigrator(sourceDBPath string, targetDB *gorm.DB, logger *log.Logger) (*DBMigrator, error) {
	// Open the source database using the mattn/go-sqlite3 driver
	sourceDB, err := sql.Open("sqlite3", sourceDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source database: %v", err)
	}

	// Check connection
	if err = sourceDB.Ping(); err != nil {
		sourceDB.Close()
		return nil, fmt.Errorf("error connecting to source database: %v", err)
	}

	return &DBMigrator{
		sourceDB: sourceDB,
		targetDB: targetDB,
		logger:   logger,
	}, nil
}

// Close closes the source database connection
func (m *DBMigrator) Close() error {
	return m.sourceDB.Close()
}

// MigrateAll migrates all data from the source database to the target database
func (m *DBMigrator) MigrateAll() error {
	// Begin a transaction on the target database
	tx := m.targetDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	// Migrate customers first
	if err := m.migrateCustomers(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Migrate accounts after customers
	if err := m.migrateAccounts(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Migrate transactions after accounts
	if err := m.migrateTransactions(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

func (m *DBMigrator) migrateCustomers(tx *gorm.DB) error {
	m.logger.Println("Migrating customers...")

	// Query all customers from the source database
	rows, err := m.sourceDB.Query(`
		SELECT id, first_name, last_name, email, phone, address 
		FROM customers
	`)
	if err != nil {
		return fmt.Errorf("error querying customers: %v", err)
	}
	defer rows.Close()

	// Migrate each customer
	var migratedCount int
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email,
			&customer.Phone,
			&customer.Address,
		)
		if err != nil {
			return fmt.Errorf("error scanning customer: %v", err)
		}

		// Create the customer in the target database
		if err := tx.Create(&customer).Error; err != nil {
			return fmt.Errorf("error creating customer in target database: %v", err)
		}

		migratedCount++
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating customer rows: %v", err)
	}

	m.logger.Printf("Successfully migrated %d customers\n", migratedCount)
	return nil
}

func (m *DBMigrator) migrateAccounts(tx *gorm.DB) error {
	m.logger.Println("Migrating accounts...")

	// Query all accounts from the source database
	rows, err := m.sourceDB.Query(`
		SELECT id, customer_id, balance, account_type, created_at, last_activity 
		FROM accounts
	`)
	if err != nil {
		return fmt.Errorf("error querying accounts: %v", err)
	}
	defer rows.Close()

	// Migrate each account
	var migratedCount int
	for rows.Next() {
		var account models.Account
		var createdAtStr, lastActivityStr string

		err := rows.Scan(
			&account.ID,
			&account.CustomerID,
			&account.Balance,
			&account.AccountType,
			&createdAtStr,
			&lastActivityStr,
		)
		if err != nil {
			return fmt.Errorf("error scanning account: %v", err)
		}

		// Parse time strings
		account.CreatedAt, err = parseTime(createdAtStr)
		if err != nil {
			return fmt.Errorf("error parsing created_at time: %v", err)
		}

		account.LastActivity, err = parseTime(lastActivityStr)
		if err != nil {
			return fmt.Errorf("error parsing last_activity time: %v", err)
		}

		// Create the account in the target database
		if err := tx.Create(&account).Error; err != nil {
			return fmt.Errorf("error creating account in target database: %v", err)
		}

		migratedCount++
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating account rows: %v", err)
	}

	m.logger.Printf("Successfully migrated %d accounts\n", migratedCount)
	return nil
}

func (m *DBMigrator) migrateTransactions(tx *gorm.DB) error {
	m.logger.Println("Migrating transactions...")

	// Query all transactions from the source database
	rows, err := m.sourceDB.Query(`
		SELECT id, account_id, amount, transaction_type, description, timestamp, destination_account_id 
		FROM transactions
	`)
	if err != nil {
		return fmt.Errorf("error querying transactions: %v", err)
	}
	defer rows.Close()

	// Migrate each transaction
	var migratedCount int
	for rows.Next() {
		var transaction models.Transaction
		var timestampStr string
		var destinationAccountID sql.NullString

		err := rows.Scan(
			&transaction.ID,
			&transaction.AccountID,
			&transaction.Amount,
			&transaction.TransactionType,
			&transaction.Description,
			&timestampStr,
			&destinationAccountID,
		)
		if err != nil {
			return fmt.Errorf("error scanning transaction: %v", err)
		}

		// Parse time string
		transaction.Timestamp, err = parseTime(timestampStr)
		if err != nil {
			return fmt.Errorf("error parsing timestamp: %v", err)
		}

		// Handle nullable destination account ID
		if destinationAccountID.Valid {
			transaction.DestinationAccountID = destinationAccountID.String
		}

		// Create the transaction in the target database
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("error creating transaction in target database: %v", err)
		}

		migratedCount++
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating transaction rows: %v", err)
	}

	m.logger.Printf("Successfully migrated %d transactions\n", migratedCount)
	return nil
}

// parseTime parses a time string in various formats
func parseTime(timeStr string) (time.Time, error) {
	// Try different time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}
