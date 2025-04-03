package models

import (
	"errors"
	"time"
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
	ID           string
	CustomerID   string
	Balance      float64
	AccountType  AccountType
	CreatedAt    time.Time
	LastActivity time.Time
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
