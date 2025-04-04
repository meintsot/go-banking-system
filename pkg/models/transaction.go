package models

import (
	"time"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	Deposit    TransactionType = "DEPOSIT"
	Withdrawal TransactionType = "WITHDRAWAL"
	Transfer   TransactionType = "TRANSFER"
)

// Transaction represents a bank transaction
type Transaction struct {
	ID                   string          `gorm:"primaryKey"`
	AccountID            string          `gorm:"index;not null"`
	Account              *Account        `gorm:"foreignKey:AccountID"`
	Amount               float64         `gorm:"not null"`
	TransactionType      TransactionType `gorm:"not null"`
	Description          string
	Timestamp            time.Time `gorm:"not null;index"`
	DestinationAccountID string    `gorm:"index"` // Used for transfers
	DestinationAccount   *Account  `gorm:"foreignKey:DestinationAccountID"`
}

// NewTransaction creates a new transaction record
func NewTransaction(id, accountID string, amount float64, transactionType TransactionType, description string) *Transaction {
	return &Transaction{
		ID:              id,
		AccountID:       accountID,
		Amount:          amount,
		TransactionType: transactionType,
		Description:     description,
		Timestamp:       time.Now(),
	}
}

// NewTransfer creates a new transfer transaction
func NewTransfer(id, sourceAccountID, destinationAccountID string, amount float64, description string) *Transaction {
	transaction := NewTransaction(id, sourceAccountID, amount, Transfer, description)
	transaction.DestinationAccountID = destinationAccountID
	return transaction
}

// GetID returns the transaction's ID to satisfy the Entity interface
func (t *Transaction) GetID() string {
	return t.ID
}
