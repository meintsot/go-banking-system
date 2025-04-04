package models

import (
	"errors"

	"gorm.io/gorm"
)

// Entity represents a generic entity with an ID
type Entity interface {
	GetID() string
}

// Customer represents a bank customer
type Customer struct {
	ID        string `gorm:"primaryKey"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	Email     string `gorm:"uniqueIndex;not null"`
	Phone     string
	Address   string
}

// NewCustomer creates a new customer with provided details
func NewCustomer(id, firstName, lastName, email, phone, address string) *Customer {
	return &Customer{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
		Address:   address,
	}
}

// FullName returns the full name of the customer
func (c *Customer) FullName() string {
	return c.FirstName + " " + c.LastName
}

// GetID returns the customer's ID to satisfy the Entity interface
func (c *Customer) GetID() string {
	return c.ID
}

// BeforeDelete is a GORM hook that checks if a customer has accounts before deletion
func (c *Customer) BeforeDelete(tx *gorm.DB) (err error) {
	var count int64
	if err := tx.Model(&Account{}).Where("customer_id = ?", c.ID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrorCustomerHasAccounts
	}
	return nil
}

// Custom errors defined for GORM hooks
var (
	ErrorCustomerHasAccounts = errors.New("cannot delete customer with active accounts")
)
