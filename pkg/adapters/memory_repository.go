package adapters

import (
	"bankingsystem/pkg/models"
	"errors"
	"sync"
)

// InMemoryCustomerRepository implements CustomerRepository using in-memory storage
type InMemoryCustomerRepository struct {
	customers map[string]*models.Customer
	mu        sync.RWMutex
}

// NewInMemoryCustomerRepository creates a new in-memory customer repository
func NewInMemoryCustomerRepository() *InMemoryCustomerRepository {
	return &InMemoryCustomerRepository{
		customers: make(map[string]*models.Customer),
	}
}

// Create adds a new customer to the repository
func (r *InMemoryCustomerRepository) Create(customer *models.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.customers[customer.ID]; exists {
		return errors.New("customer with this ID already exists")
	}

	r.customers[customer.ID] = customer
	return nil
}

// GetByID retrieves a customer by their ID
func (r *InMemoryCustomerRepository) GetByID(id string) (*models.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	customer, exists := r.customers[id]
	if !exists {
		return nil, nil // Not found, but not an error
	}

	return customer, nil
}

// Update updates an existing customer
func (r *InMemoryCustomerRepository) Update(customer *models.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.customers[customer.ID]; !exists {
		return errors.New("customer not found")
	}

	r.customers[customer.ID] = customer
	return nil
}

// Delete removes a customer by their ID
func (r *InMemoryCustomerRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.customers[id]; !exists {
		return errors.New("customer not found")
	}

	delete(r.customers, id)
	return nil
}

// List returns all customers
func (r *InMemoryCustomerRepository) List() ([]*models.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	customers := make([]*models.Customer, 0, len(r.customers))
	for _, customer := range r.customers {
		customers = append(customers, customer)
	}

	return customers, nil
}

// InMemoryAccountRepository implements AccountRepository using in-memory storage
type InMemoryAccountRepository struct {
	accounts map[string]*models.Account
	mu       sync.RWMutex
}

// NewInMemoryAccountRepository creates a new in-memory account repository
func NewInMemoryAccountRepository() *InMemoryAccountRepository {
	return &InMemoryAccountRepository{
		accounts: make(map[string]*models.Account),
	}
}

// Create adds a new account to the repository
func (r *InMemoryAccountRepository) Create(account *models.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[account.ID]; exists {
		return errors.New("account with this ID already exists")
	}

	r.accounts[account.ID] = account
	return nil
}

// GetByID retrieves an account by its ID
func (r *InMemoryAccountRepository) GetByID(id string) (*models.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, exists := r.accounts[id]
	if !exists {
		return nil, nil // Not found, but not an error
	}

	return account, nil
}

// GetByCustomerID retrieves all accounts for a customer
func (r *InMemoryAccountRepository) GetByCustomerID(customerID string) ([]*models.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var accounts []*models.Account
	for _, account := range r.accounts {
		if account.CustomerID == customerID {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}

// Update updates an existing account
func (r *InMemoryAccountRepository) Update(account *models.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[account.ID]; !exists {
		return errors.New("account not found")
	}

	r.accounts[account.ID] = account
	return nil
}

// Delete removes an account by its ID
func (r *InMemoryAccountRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.accounts[id]; !exists {
		return errors.New("account not found")
	}

	delete(r.accounts, id)
	return nil
}

// InMemoryTransactionRepository implements TransactionRepository using in-memory storage
type InMemoryTransactionRepository struct {
	transactions map[string]*models.Transaction
	mu           sync.RWMutex
}

// NewInMemoryTransactionRepository creates a new in-memory transaction repository
func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[string]*models.Transaction),
	}
}

// Create adds a new transaction to the repository
func (r *InMemoryTransactionRepository) Create(transaction *models.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transactions[transaction.ID]; exists {
		return errors.New("transaction with this ID already exists")
	}

	r.transactions[transaction.ID] = transaction
	return nil
}

// GetByID retrieves a transaction by its ID
func (r *InMemoryTransactionRepository) GetByID(id string) (*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transaction, exists := r.transactions[id]
	if !exists {
		return nil, nil // Not found, but not an error
	}

	return transaction, nil
}

// GetByAccountID retrieves all transactions for an account
func (r *InMemoryTransactionRepository) GetByAccountID(accountID string) ([]*models.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var transactions []*models.Transaction
	for _, transaction := range r.transactions {
		if transaction.AccountID == accountID || transaction.DestinationAccountID == accountID {
			transactions = append(transactions, transaction)
		}
	}

	return transactions, nil
}
