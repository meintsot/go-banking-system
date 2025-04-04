package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// AccountType represents the type of bank account
type AccountType string

const (
	Checking AccountType = "CHECKING"
	Savings  AccountType = "SAVINGS"
	Business AccountType = "BUSINESS"
)

// Account represents a bank account with common fields
type Account struct {
	ID           string        `gorm:"primaryKey"`
	CustomerID   string        `gorm:"index;not null"`
	Customer     *Customer     `gorm:"foreignKey:CustomerID"`
	Balance      float64       `gorm:"not null"`
	AccountType  AccountType   `gorm:"not null"`
	CreatedAt    time.Time     `gorm:"not null"`
	LastActivity time.Time     `gorm:"not null"`
	Transactions []Transaction `gorm:"foreignKey:AccountID"`
}

// NewAccount creates a new account with the provided details
func NewAccount(id, customerID string, initialBalance float64, accountType AccountType) *Account {
	now := time.Now()
	return &Account{
		ID:           id,
		CustomerID:   customerID,
		Balance:      initialBalance,
		AccountType:  accountType,
		CreatedAt:    now,
		LastActivity: now,
	}
}

// Deposit adds the specified amount to the account balance
func (a *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}

	a.Balance += amount
	a.LastActivity = time.Now()
	return nil
}

// Withdraw subtracts the specified amount from the account balance
func (a *Account) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}

	if a.Balance < amount {
		return errors.New("insufficient funds")
	}

	a.Balance -= amount
	a.LastActivity = time.Now()
	return nil
}

// GetBalance returns the current balance
func (a *Account) GetBalance() float64 {
	return a.Balance
}

// GetID returns the account's ID to satisfy the Entity interface
func (a *Account) GetID() string {
	return a.ID
}

// BeforeDelete is a GORM hook that checks if an account has a positive balance before deletion
func (a *Account) BeforeDelete(tx *gorm.DB) (err error) {
	if a.Balance > 0 {
		return ErrorAccountHasBalance
	}
	return nil
}

// Custom errors for GORM hooks
var (
	ErrorAccountHasBalance = errors.New("cannot delete account with positive balance")
)
