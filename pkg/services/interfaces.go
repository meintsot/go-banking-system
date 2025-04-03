package services

import (
	"bankingsystem/pkg/models"
)

// CustomerRepository defines operations for customer data management
type CustomerRepository interface {
	Create(customer *models.Customer) error
	GetByID(id string) (*models.Customer, error)
	Update(customer *models.Customer) error
	Delete(id string) error
	List() ([]*models.Customer, error)
}

// AccountRepository defines operations for account data management
type AccountRepository interface {
	Create(account *models.Account) error
	GetByID(id string) (*models.Account, error)
	GetByCustomerID(customerID string) ([]*models.Account, error)
	Update(account *models.Account) error
	Delete(id string) error
}

// TransactionRepository defines operations for transaction data management
type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	GetByID(id string) (*models.Transaction, error)
	GetByAccountID(accountID string) ([]*models.Transaction, error)
}

// BankingService defines the main operations of the banking system
type BankingService interface {
	CreateCustomer(firstName, lastName, email, phone, address string) (*models.Customer, error)
	GetCustomer(id string) (*models.Customer, error)
	UpdateCustomer(customer *models.Customer) error
	DeleteCustomer(id string) error
	ListCustomers() ([]*models.Customer, error)

	CreateAccount(customerID string, initialBalance float64, accountType models.AccountType) (*models.Account, error)
	GetAccount(id string) (*models.Account, error)
	GetCustomerAccounts(customerID string) ([]*models.Account, error)
	DeleteAccount(id string) error

	Deposit(accountID string, amount float64) error
	Withdraw(accountID string, amount float64) error
	Transfer(sourceAccountID, destinationAccountID string, amount float64) error

	GetAccountTransactions(accountID string) ([]*models.Transaction, error)
}
