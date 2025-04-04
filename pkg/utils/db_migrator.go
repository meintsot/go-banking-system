package utils

import (
	"bankingsystem/pkg/models"
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite" // Pure Go SQLite driver for GORM
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DBMigrator is a utility for migrating data from old SQLite database to GORM database
type DBMigrator struct {
	sourceDB *gorm.DB
	targetDB *gorm.DB
	logger   *log.Logger
}

// NewDBMigrator creates a new database migrator
func NewDBMigrator(sourceDBPath string, targetDB *gorm.DB, logger *log.Logger) (*DBMigrator, error) {
	// Set up GORM configuration for source DB
	config := &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(gormlogger.Error), // Silence most GORM logs for source DB
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// Open source database connection with GORM
	sourceDB, err := gorm.Open(sqlite.Open(sourceDBPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open source database with GORM: %v", err)
	}

	return &DBMigrator{
		sourceDB: sourceDB,
		targetDB: targetDB,
		logger:   logger,
	}, nil
}

// MigrateAll migrates all data from the source database to the target database
func (m *DBMigrator) MigrateAll() error {
	// First, let's check what tables are available in the source database
	var tables []string
	if err := m.sourceDB.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables).Error; err != nil {
		return fmt.Errorf("error checking tables in source database: %v", err)
	}

	m.logger.Printf("Found tables in source database: %v", tables)

	// For each table, print the schema
	for _, table := range tables {
		var schema string
		if err := m.sourceDB.Raw("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&schema).Error; err != nil {
			m.logger.Printf("Warning: Could not get schema for table %s: %v", table, err)
			continue
		}
		m.logger.Printf("Schema for table %s: %s", table, schema)

		// Get row count for this table
		var count int64
		if err := m.sourceDB.Table(table).Count(&count).Error; err != nil {
			m.logger.Printf("Warning: Could not count rows in table %s: %v", table, err)
			continue
		}
		m.logger.Printf("Table %s has %d rows", table, count)
	}

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

	// Create temporary model structs for source DB that match old schema
	type OldCustomer struct {
		ID        string
		FirstName string `gorm:"column:first_name"`
		LastName  string `gorm:"column:last_name"`
		Email     string
		Phone     string
		Address   string
	}

	// Tell GORM the table name explicitly
	m.sourceDB.Table("customers")

	// Read all customers from source database
	var oldCustomers []OldCustomer
	if err := m.sourceDB.Table("customers").Find(&oldCustomers).Error; err != nil {
		return fmt.Errorf("error querying customers: %v", err)
	}

	// Migrate each customer to the target database
	for _, oldCustomer := range oldCustomers {
		newCustomer := models.Customer{
			ID:        oldCustomer.ID,
			FirstName: oldCustomer.FirstName,
			LastName:  oldCustomer.LastName,
			Email:     oldCustomer.Email,
			Phone:     oldCustomer.Phone,
			Address:   oldCustomer.Address,
		}

		if err := tx.Create(&newCustomer).Error; err != nil {
			return fmt.Errorf("error creating customer in target database: %v", err)
		}
	}

	m.logger.Printf("Successfully migrated %d customers\n", len(oldCustomers))
	return nil
}

func (m *DBMigrator) migrateAccounts(tx *gorm.DB) error {
	m.logger.Println("Migrating accounts...")

	// Create temporary model structs for source DB that match old schema
	type OldAccount struct {
		ID           string
		CustomerID   string `gorm:"column:customer_id"`
		Balance      float64
		AccountType  string `gorm:"column:account_type"`
		CreatedAt    string `gorm:"column:created_at"`
		LastActivity string `gorm:"column:last_activity"`
	}

	// Read all accounts from source database
	var oldAccounts []OldAccount
	if err := m.sourceDB.Table("accounts").Find(&oldAccounts).Error; err != nil {
		return fmt.Errorf("error querying accounts: %v", err)
	}

	// Migrate each account to the target database
	for _, oldAccount := range oldAccounts {
		// Parse time fields
		createdAt, err := parseTime(oldAccount.CreatedAt)
		if err != nil {
			m.logger.Printf("Warning: Could not parse created_at time for account %s: %v - using current time", oldAccount.ID, err)
			createdAt = time.Now()
		}

		lastActivity, err := parseTime(oldAccount.LastActivity)
		if err != nil {
			m.logger.Printf("Warning: Could not parse last_activity time for account %s: %v - using current time", oldAccount.ID, err)
			lastActivity = time.Now()
		}

		newAccount := models.Account{
			ID:           oldAccount.ID,
			CustomerID:   oldAccount.CustomerID,
			Balance:      oldAccount.Balance,
			AccountType:  models.AccountType(oldAccount.AccountType),
			CreatedAt:    createdAt,
			LastActivity: lastActivity,
		}

		if err := tx.Create(&newAccount).Error; err != nil {
			return fmt.Errorf("error creating account in target database: %v", err)
		}
	}

	m.logger.Printf("Successfully migrated %d accounts\n", len(oldAccounts))
	return nil
}

func (m *DBMigrator) migrateTransactions(tx *gorm.DB) error {
	m.logger.Println("Migrating transactions...")

	// Create temporary model structs for source DB that match old schema
	type OldTransaction struct {
		ID                   string
		AccountID            string `gorm:"column:account_id"`
		Amount               float64
		TransactionType      string `gorm:"column:transaction_type"`
		Description          string
		Timestamp            string
		DestinationAccountID string `gorm:"column:destination_account_id"`
	}

	// Read all transactions from source database
	var oldTransactions []OldTransaction
	if err := m.sourceDB.Table("transactions").Find(&oldTransactions).Error; err != nil {
		return fmt.Errorf("error querying transactions: %v", err)
	}

	// Migrate each transaction to the target database
	for _, oldTx := range oldTransactions {
		// Parse timestamp
		timestamp, err := parseTime(oldTx.Timestamp)
		if err != nil {
			m.logger.Printf("Warning: Could not parse timestamp for transaction %s: %v - using current time", oldTx.ID, err)
			timestamp = time.Now()
		}

		newTransaction := models.Transaction{
			ID:                   oldTx.ID,
			AccountID:            oldTx.AccountID,
			Amount:               oldTx.Amount,
			TransactionType:      models.TransactionType(oldTx.TransactionType),
			Description:          oldTx.Description,
			Timestamp:            timestamp,
			DestinationAccountID: oldTx.DestinationAccountID,
		}

		if err := tx.Create(&newTransaction).Error; err != nil {
			return fmt.Errorf("error creating transaction in target database: %v", err)
		}
	}

	m.logger.Printf("Successfully migrated %d transactions\n", len(oldTransactions))
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

// Close closes the database connections
func (m *DBMigrator) Close() error {
	if m.sourceDB != nil {
		sqlDB, err := m.sourceDB.DB()
		if err != nil {
			return fmt.Errorf("error getting source SQL DB: %v", err)
		}
		return sqlDB.Close()
	}
	return nil
}
