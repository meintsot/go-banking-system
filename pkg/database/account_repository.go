package database

import (
	"bankingsystem/pkg/models"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SQLiteAccountRepository implements account repository using SQLite
type SQLiteAccountRepository struct {
	db *sql.DB
}

// NewSQLiteAccountRepository creates a new SQLite account repository
func NewSQLiteAccountRepository(conn *DBConnection) *SQLiteAccountRepository {
	return &SQLiteAccountRepository{
		db: conn.GetDB(),
	}
}

// Create inserts a new account
func (r *SQLiteAccountRepository) Create(account *models.Account) error {
	query := `INSERT INTO accounts (id, customer_id, balance, account_type, created_at, last_activity) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		account.ID,
		account.CustomerID,
		account.Balance,
		account.AccountType,
		account.CreatedAt.Format(time.RFC3339),
		account.LastActivity.Format(time.RFC3339))

	if err != nil {
		return fmt.Errorf("error creating account: %v", err)
	}

	return nil
}

// GetByID retrieves an account by ID
func (r *SQLiteAccountRepository) GetByID(id string) (*models.Account, error) {
	query := `SELECT id, customer_id, balance, account_type, created_at, last_activity 
			  FROM accounts WHERE id = ?`

	var account models.Account
	var createdAtStr, lastActivityStr string

	err := r.db.QueryRow(query, id).Scan(
		&account.ID,
		&account.CustomerID,
		&account.Balance,
		&account.AccountType,
		&createdAtStr,
		&lastActivityStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("error getting account: %v", err)
	}

	// Parse time strings
	account.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %v", err)
	}

	account.LastActivity, err = time.Parse(time.RFC3339, lastActivityStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing last_activity: %v", err)
	}

	return &account, nil
}

// GetByCustomerID retrieves all accounts for a customer
func (r *SQLiteAccountRepository) GetByCustomerID(customerID string) ([]*models.Account, error) {
	query := `SELECT id, customer_id, balance, account_type, created_at, last_activity 
			  FROM accounts WHERE customer_id = ?`

	rows, err := r.db.Query(query, customerID)
	if err != nil {
		return nil, fmt.Errorf("error getting customer accounts: %v", err)
	}
	defer rows.Close()

	accounts := []*models.Account{}

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
			return nil, fmt.Errorf("error scanning account row: %v", err)
		}

		// Parse time strings
		account.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at: %v", err)
		}

		account.LastActivity, err = time.Parse(time.RFC3339, lastActivityStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing last_activity: %v", err)
		}

		accounts = append(accounts, &account)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating account rows: %v", err)
	}

	return accounts, nil
}

// Update updates an account
func (r *SQLiteAccountRepository) Update(account *models.Account) error {
	query := `UPDATE accounts 
			  SET balance = ?, account_type = ?, last_activity = ? 
			  WHERE id = ?`

	result, err := r.db.Exec(query,
		account.Balance,
		account.AccountType,
		account.LastActivity.Format(time.RFC3339),
		account.ID)

	if err != nil {
		return fmt.Errorf("error updating account: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return errors.New("account not found")
	}

	return nil
}

// Delete deletes an account
func (r *SQLiteAccountRepository) Delete(id string) error {
	// Start a transaction for account deletion
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check account balance
	var balance float64
	err = tx.QueryRow("SELECT balance FROM accounts WHERE id = ?", id).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("account not found")
		}
		return fmt.Errorf("error checking account balance: %v", err)
	}

	if balance > 0 {
		return errors.New("cannot delete account with positive balance")
	}

	// Delete associated transactions
	_, err = tx.Exec("DELETE FROM transactions WHERE account_id = ? OR destination_account_id = ?", id, id)
	if err != nil {
		return fmt.Errorf("error deleting account transactions: %v", err)
	}

	// Delete the account
	result, err := tx.Exec("DELETE FROM accounts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error deleting account: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return errors.New("account not found")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
