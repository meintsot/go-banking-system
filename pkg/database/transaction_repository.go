package database

import (
	"bankingsystem/pkg/models"
	"database/sql"
	"fmt"
	"time"
)

// SQLiteTransactionRepository implements transaction repository using SQLite
type SQLiteTransactionRepository struct {
	db *sql.DB
}

// NewSQLiteTransactionRepository creates a new SQLite transaction repository
func NewSQLiteTransactionRepository(conn *DBConnection) *SQLiteTransactionRepository {
	return &SQLiteTransactionRepository{
		db: conn.GetDB(),
	}
}

// Create inserts a new transaction
func (r *SQLiteTransactionRepository) Create(transaction *models.Transaction) error {
	query := `INSERT INTO transactions 
			  (id, account_id, amount, transaction_type, description, timestamp, destination_account_id) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		transaction.ID,
		transaction.AccountID,
		transaction.Amount,
		transaction.TransactionType,
		transaction.Description,
		transaction.Timestamp.Format(time.RFC3339),
		transaction.DestinationAccountID)

	if err != nil {
		return fmt.Errorf("error creating transaction: %v", err)
	}

	return nil
}

// GetByAccountID retrieves all transactions for an account
func (r *SQLiteTransactionRepository) GetByAccountID(accountID string) ([]*models.Transaction, error) {
	query := `SELECT id, account_id, amount, transaction_type, description, timestamp, destination_account_id
			  FROM transactions 
			  WHERE account_id = ? OR destination_account_id = ?
			  ORDER BY timestamp DESC`

	rows, err := r.db.Query(query, accountID, accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting account transactions: %v", err)
	}
	defer rows.Close()

	transactions := []*models.Transaction{}

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
			return nil, fmt.Errorf("error scanning transaction row: %v", err)
		}

		// Parse time string
		transaction.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing timestamp: %v", err)
		}

		// Handle nullable destination account ID
		if destinationAccountID.Valid {
			transaction.DestinationAccountID = destinationAccountID.String
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %v", err)
	}

	return transactions, nil
}

// GetByID retrieves a transaction by ID
func (r *SQLiteTransactionRepository) GetByID(id string) (*models.Transaction, error) {
	query := `SELECT id, account_id, amount, transaction_type, description, timestamp, destination_account_id
			  FROM transactions WHERE id = ?`

	var transaction models.Transaction
	var timestampStr string
	var destinationAccountID sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.AccountID,
		&transaction.Amount,
		&transaction.TransactionType,
		&transaction.Description,
		&timestampStr,
		&destinationAccountID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("error getting transaction: %v", err)
	}

	// Parse time string
	transaction.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing timestamp: %v", err)
	}

	// Handle nullable destination account ID
	if destinationAccountID.Valid {
		transaction.DestinationAccountID = destinationAccountID.String
	}

	return &transaction, nil
}

// ExecuteTransaction performs a financial transaction within a database transaction
func (r *SQLiteTransactionRepository) ExecuteTransaction(transaction *models.Transaction) error {
	// Begin database transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update source account balance
	sourceQuery := "UPDATE accounts SET balance = balance + ?, last_activity = ? WHERE id = ?"
	_, err = tx.Exec(sourceQuery,
		// For withdrawals and transfers, amount is negative
		transaction.Amount,
		transaction.Timestamp.Format(time.RFC3339),
		transaction.AccountID)

	if err != nil {
		return fmt.Errorf("error updating source account: %v", err)
	}

	// For transfers, update destination account balance
	if transaction.TransactionType == models.Transfer && transaction.DestinationAccountID != "" {
		destQuery := "UPDATE accounts SET balance = balance + ?, last_activity = ? WHERE id = ?"
		_, err = tx.Exec(destQuery,
			// For destination account, amount is positive
			transaction.Amount*-1, // Negate the amount since we stored it as negative
			transaction.Timestamp.Format(time.RFC3339),
			transaction.DestinationAccountID)
	}

	// Save the transaction record
	transQuery := `INSERT INTO transactions 
				  (id, account_id, amount, transaction_type, description, timestamp, destination_account_id) 
				  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(transQuery,
		transaction.ID,
		transaction.AccountID,
		transaction.Amount,
		transaction.TransactionType,
		transaction.Description,
		transaction.Timestamp.Format(time.RFC3339),
		transaction.DestinationAccountID)

	if err != nil {
		return fmt.Errorf("error saving transaction record: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
