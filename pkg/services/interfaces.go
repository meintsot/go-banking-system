package services

import (
	"bankingsystem/pkg/models"
)

// CustomerRepository defines the operations for managing customers
type CustomerRepository interface {
	Create(customer *models.Customer) error
	GetByID(id string) (*models.Customer, error)
	Update(customer *models.Customer) error
	Delete(id string) error
	List() ([]*models.Customer, error)
}

// AccountRepository defines the operations for managing accounts
type AccountRepository interface {
	Create(account *models.Account) error
	GetByID(id string) (*models.Account, error)
	GetByCustomerID(customerID string) ([]*models.Account, error)
	Update(account *models.Account) error
	Delete(id string) error
}

// TransactionRepository defines the operations for managing transactions
type TransactionRepository interface {
	Create(transaction *models.Transaction) error
	GetByID(id string) (*models.Transaction, error)
	GetByAccountID(accountID string) ([]*models.Transaction, error)
}

// IDGenerator defines an interface for generating unique IDs
type IDGenerator interface {
	GenerateID() string
}

// BankingServiceInterface defines the main operations of the banking system
type BankingServiceInterface interface {
	CreateCustomer(firstName, lastName, email, phone, address string) (*models.Customer, error)
	GetCustomer(id string) (*models.Customer, error)
	UpdateCustomer(customer *models.Customer) error
	DeleteCustomer(id string) error
	ListCustomers() ([]*models.Customer, error)

	CreateAccount(customerID string, accountType string, initialBalance float64) (*models.Account, error)
	GetAccount(id string) (*models.Account, error)
	GetCustomerAccounts(customerID string) ([]*models.Account, error)
	DeleteAccount(id string) error

	Deposit(accountID string, amount float64, description string) (*models.Transaction, error)
	Withdraw(accountID string, amount float64, description string) (*models.Transaction, error)
	Transfer(sourceAccountID, destinationAccountID string, amount float64, description string) (*models.Transaction, error)

	GetAccountTransactions(accountID string) ([]*models.Transaction, error)
	GetTransaction(id string) (*models.Transaction, error)
}
