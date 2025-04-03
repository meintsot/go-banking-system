package services

import (
	"bankingsystem/pkg/models"
	"fmt"
	"time"
)

// BankingService implements banking operations using repositories
type BankingService struct {
	customerRepo    CustomerRepository
	accountRepo     AccountRepository
	transactionRepo TransactionRepository
	idGenerator     IDGenerator
}

// NewBankingService creates a new banking service
func NewBankingService(
	customerRepo CustomerRepository,
	accountRepo AccountRepository,
	transactionRepo TransactionRepository,
	idGenerator IDGenerator,
) BankingServiceInterface {
	return &BankingService{
		customerRepo:    customerRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		idGenerator:     idGenerator,
	}
}

// CreateCustomer creates a new customer
func (s *BankingService) CreateCustomer(firstName, lastName, email, phone, address string) (*models.Customer, error) {
	customer := &models.Customer{
		ID:        s.idGenerator.GenerateID(),
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
		Address:   address,
	}

	err := s.customerRepo.Create(customer)
	if err != nil {
		return nil, fmt.Errorf("error creating customer: %w", err)
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *BankingService) GetCustomer(id string) (*models.Customer, error) {
	customer, err := s.customerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer with ID %s not found", id)
	}
	return customer, nil
}

// UpdateCustomer updates an existing customer
func (s *BankingService) UpdateCustomer(customer *models.Customer) error {
	// Check if customer exists
	existingCustomer, err := s.customerRepo.GetByID(customer.ID)
	if err != nil {
		return fmt.Errorf("error checking customer existence: %w", err)
	}
	if existingCustomer == nil {
		return fmt.Errorf("customer with ID %s not found", customer.ID)
	}

	err = s.customerRepo.Update(customer)
	if err != nil {
		return fmt.Errorf("error updating customer: %w", err)
	}

	return nil
}

// DeleteCustomer deletes a customer
func (s *BankingService) DeleteCustomer(id string) error {
	err := s.customerRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting customer: %w", err)
	}
	return nil
}

// ListCustomers returns all customers
func (s *BankingService) ListCustomers() ([]*models.Customer, error) {
	customers, err := s.customerRepo.List()
	if err != nil {
		return nil, fmt.Errorf("error listing customers: %w", err)
	}
	return customers, nil
}

// CreateAccount creates a new account for a customer
func (s *BankingService) CreateAccount(customerID, accountType string, initialDeposit float64) (*models.Account, error) {
	// Check if customer exists
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, fmt.Errorf("error checking customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer with ID %s not found", customerID)
	}

	// Validate initial deposit
	if initialDeposit < 0 {
		return nil, fmt.Errorf("initial deposit cannot be negative")
	}

	now := time.Now()
	account := &models.Account{
		ID:           s.idGenerator.GenerateID(),
		CustomerID:   customerID,
		Balance:      initialDeposit,
		AccountType:  models.AccountType(accountType),
		CreatedAt:    now,
		LastActivity: now,
	}

	// Create the account
	err = s.accountRepo.Create(account)
	if err != nil {
		return nil, fmt.Errorf("error creating account: %w", err)
	}

	// Record initial deposit transaction if > 0
	if initialDeposit > 0 {
		transaction := &models.Transaction{
			ID:              s.idGenerator.GenerateID(),
			AccountID:       account.ID,
			Amount:          initialDeposit,
			TransactionType: models.Deposit,
			Description:     "Initial deposit",
			Timestamp:       now,
		}

		err = s.transactionRepo.Create(transaction)
		if err != nil {
			// Log this error but don't fail the account creation
			// In a production system, this should be handled with a transaction
			fmt.Printf("Warning: Failed to record initial deposit: %v\n", err)
		}
	}

	return account, nil
}

// GetAccount retrieves an account by ID
func (s *BankingService) GetAccount(id string) (*models.Account, error) {
	account, err := s.accountRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account with ID %s not found", id)
	}
	return account, nil
}

// GetCustomerAccounts retrieves all accounts for a customer
func (s *BankingService) GetCustomerAccounts(customerID string) ([]*models.Account, error) {
	accounts, err := s.accountRepo.GetByCustomerID(customerID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving customer accounts: %w", err)
	}
	return accounts, nil
}

// DeleteAccount deletes an account
func (s *BankingService) DeleteAccount(id string) error {
	err := s.accountRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("error deleting account: %w", err)
	}
	return nil
}

// Deposit adds funds to an account
func (s *BankingService) Deposit(accountID string, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("deposit amount must be positive")
	}

	// Get the account
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account with ID %s not found", accountID)
	}

	now := time.Now()

	// Update account balance
	account.Balance += amount
	account.LastActivity = now
	err = s.accountRepo.Update(account)
	if err != nil {
		return nil, fmt.Errorf("error updating account balance: %w", err)
	}

	// Record the transaction
	transaction := &models.Transaction{
		ID:              s.idGenerator.GenerateID(),
		AccountID:       accountID,
		Amount:          amount,
		TransactionType: models.Deposit,
		Description:     description,
		Timestamp:       now,
	}

	err = s.transactionRepo.Create(transaction)
	if err != nil {
		// This should ideally be in a database transaction to ensure atomicity
		return nil, fmt.Errorf("error recording transaction: %w", err)
	}

	return transaction, nil
}

// Withdraw removes funds from an account
func (s *BankingService) Withdraw(accountID string, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("withdrawal amount must be positive")
	}

	// Get the account
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account with ID %s not found", accountID)
	}

	// Check if sufficient funds
	if account.Balance < amount {
		return nil, fmt.Errorf("insufficient funds")
	}

	now := time.Now()

	// Update account balance
	account.Balance -= amount
	account.LastActivity = now
	err = s.accountRepo.Update(account)
	if err != nil {
		return nil, fmt.Errorf("error updating account balance: %w", err)
	}

	// Record the transaction
	transaction := &models.Transaction{
		ID:              s.idGenerator.GenerateID(),
		AccountID:       accountID,
		Amount:          amount,
		TransactionType: models.Withdrawal,
		Description:     description,
		Timestamp:       now,
	}

	err = s.transactionRepo.Create(transaction)
	if err != nil {
		// This should ideally be in a database transaction to ensure atomicity
		return nil, fmt.Errorf("error recording transaction: %w", err)
	}

	return transaction, nil
}

// Transfer moves funds from one account to another
func (s *BankingService) Transfer(
	sourceAccountID string,
	destinationAccountID string,
	amount float64,
	description string,
) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("transfer amount must be positive")
	}

	// Get the source account
	sourceAccount, err := s.accountRepo.GetByID(sourceAccountID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving source account: %w", err)
	}
	if sourceAccount == nil {
		return nil, fmt.Errorf("source account with ID %s not found", sourceAccountID)
	}

	// Check if sufficient funds
	if sourceAccount.Balance < amount {
		return nil, fmt.Errorf("insufficient funds in source account")
	}

	// Get the destination account
	destAccount, err := s.accountRepo.GetByID(destinationAccountID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving destination account: %w", err)
	}
	if destAccount == nil {
		return nil, fmt.Errorf("destination account with ID %s not found", destinationAccountID)
	}

	now := time.Now()

	// Update source account
	sourceAccount.Balance -= amount
	sourceAccount.LastActivity = now
	err = s.accountRepo.Update(sourceAccount)
	if err != nil {
		return nil, fmt.Errorf("error updating source account: %w", err)
	}

	// Update destination account
	destAccount.Balance += amount
	destAccount.LastActivity = now
	err = s.accountRepo.Update(destAccount)
	if err != nil {
		// In a production system, this should be handled with a database transaction
		// to ensure atomicity and rollback capability
		return nil, fmt.Errorf("error updating destination account: %w", err)
	}

	// Record the transaction
	transaction := &models.Transaction{
		ID:                   s.idGenerator.GenerateID(),
		AccountID:            sourceAccountID,
		Amount:               amount,
		TransactionType:      models.Transfer,
		Description:          description,
		Timestamp:            now,
		DestinationAccountID: destinationAccountID,
	}

	err = s.transactionRepo.Create(transaction)
	if err != nil {
		// This should ideally be in a database transaction to ensure atomicity
		return nil, fmt.Errorf("error recording transaction: %w", err)
	}

	return transaction, nil
}

// GetAccountTransactions retrieves all transactions for an account
func (s *BankingService) GetAccountTransactions(accountID string) ([]*models.Transaction, error) {
	// Verify the account exists
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("error verifying account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account with ID %s not found", accountID)
	}

	transactions, err := s.transactionRepo.GetByAccountID(accountID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving account transactions: %w", err)
	}
	return transactions, nil
}

// GetTransaction retrieves a transaction by ID
func (s *BankingService) GetTransaction(id string) (*models.Transaction, error) {
	transaction, err := s.transactionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving transaction: %w", err)
	}
	if transaction == nil {
		return nil, fmt.Errorf("transaction with ID %s not found", id)
	}
	return transaction, nil
}
