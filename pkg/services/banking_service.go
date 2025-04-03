package services

import (
	"bankingsystem/pkg/models"
	"errors"
	"fmt"
)

// DefaultBankingService is the default implementation of BankingService
type DefaultBankingService struct {
	customerRepo    CustomerRepository
	accountRepo     AccountRepository
	transactionRepo TransactionRepository
	idGenerator     IDGenerator
}

// IDGenerator interface for generating unique IDs
type IDGenerator interface {
	GenerateID() string
}

// NewBankingService creates a new banking service
func NewBankingService(
	customerRepo CustomerRepository,
	accountRepo AccountRepository,
	transactionRepo TransactionRepository,
	idGenerator IDGenerator,
) *DefaultBankingService {
	return &DefaultBankingService{
		customerRepo:    customerRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		idGenerator:     idGenerator,
	}
}

// CreateCustomer creates a new customer
func (s *DefaultBankingService) CreateCustomer(firstName, lastName, email, phone, address string) (*models.Customer, error) {
	id := s.idGenerator.GenerateID()
	customer := models.NewCustomer(id, firstName, lastName, email, phone, address)

	err := s.customerRepo.Create(customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *DefaultBankingService) GetCustomer(id string) (*models.Customer, error) {
	return s.customerRepo.GetByID(id)
}

// UpdateCustomer updates an existing customer
func (s *DefaultBankingService) UpdateCustomer(customer *models.Customer) error {
	return s.customerRepo.Update(customer)
}

// DeleteCustomer deletes a customer by ID
func (s *DefaultBankingService) DeleteCustomer(id string) error {
	accounts, err := s.accountRepo.GetByCustomerID(id)
	if err != nil {
		return err
	}

	if len(accounts) > 0 {
		return errors.New("cannot delete customer with active accounts")
	}

	return s.customerRepo.Delete(id)
}

// ListCustomers lists all customers
func (s *DefaultBankingService) ListCustomers() ([]*models.Customer, error) {
	return s.customerRepo.List()
}

// CreateAccount creates a new account for a customer
func (s *DefaultBankingService) CreateAccount(customerID string, initialBalance float64, accountType models.AccountType) (*models.Account, error) {
	// Verify customer exists
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, err
	}

	if customer == nil {
		return nil, errors.New("customer not found")
	}

	id := s.idGenerator.GenerateID()
	account := models.NewAccount(id, customerID, initialBalance, accountType)

	err = s.accountRepo.Create(account)
	if err != nil {
		return nil, err
	}

	// Record initial deposit if balance > 0
	if initialBalance > 0 {
		transaction := models.NewTransaction(
			s.idGenerator.GenerateID(),
			account.ID,
			initialBalance,
			models.Deposit,
			"Initial deposit",
		)

		err = s.transactionRepo.Create(transaction)
		if err != nil {
			return nil, fmt.Errorf("account created but failed to record initial deposit: %w", err)
		}
	}

	return account, nil
}

// GetAccount retrieves an account by ID
func (s *DefaultBankingService) GetAccount(id string) (*models.Account, error) {
	return s.accountRepo.GetByID(id)
}

// GetCustomerAccounts retrieves all accounts for a customer
func (s *DefaultBankingService) GetCustomerAccounts(customerID string) ([]*models.Account, error) {
	return s.accountRepo.GetByCustomerID(customerID)
}

// DeleteAccount deletes an account by ID
func (s *DefaultBankingService) DeleteAccount(id string) error {
	account, err := s.accountRepo.GetByID(id)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("account not found")
	}

	if account.Balance > 0 {
		return errors.New("cannot delete account with positive balance")
	}

	return s.accountRepo.Delete(id)
}

// Deposit adds amount to an account
func (s *DefaultBankingService) Deposit(accountID string, amount float64) error {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("account not found")
	}

	err = account.Deposit(amount)
	if err != nil {
		return err
	}

	err = s.accountRepo.Update(account)
	if err != nil {
		return err
	}

	// Record transaction
	transaction := models.NewTransaction(
		s.idGenerator.GenerateID(),
		accountID,
		amount,
		models.Deposit,
		"Deposit",
	)

	return s.transactionRepo.Create(transaction)
}

// Withdraw removes amount from an account
func (s *DefaultBankingService) Withdraw(accountID string, amount float64) error {
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("account not found")
	}

	err = account.Withdraw(amount)
	if err != nil {
		return err
	}

	err = s.accountRepo.Update(account)
	if err != nil {
		return err
	}

	// Record transaction
	transaction := models.NewTransaction(
		s.idGenerator.GenerateID(),
		accountID,
		amount,
		models.Withdrawal,
		"Withdrawal",
	)

	return s.transactionRepo.Create(transaction)
}

// Transfer moves amount from one account to another
func (s *DefaultBankingService) Transfer(sourceAccountID, destinationAccountID string, amount float64) error {
	if sourceAccountID == destinationAccountID {
		return errors.New("source and destination accounts cannot be the same")
	}

	sourceAccount, err := s.accountRepo.GetByID(sourceAccountID)
	if err != nil {
		return err
	}

	if sourceAccount == nil {
		return errors.New("source account not found")
	}

	destinationAccount, err := s.accountRepo.GetByID(destinationAccountID)
	if err != nil {
		return err
	}

	if destinationAccount == nil {
		return errors.New("destination account not found")
	}

	// Withdraw from source
	err = sourceAccount.Withdraw(amount)
	if err != nil {
		return err
	}

	// Deposit to destination
	err = destinationAccount.Deposit(amount)
	if err != nil {
		// Rollback withdrawal
		sourceAccount.Deposit(amount)
		return err
	}

	// Update both accounts
	err = s.accountRepo.Update(sourceAccount)
	if err != nil {
		// Rollback
		sourceAccount.Deposit(amount)
		destinationAccount.Withdraw(amount)
		return err
	}

	err = s.accountRepo.Update(destinationAccount)
	if err != nil {
		// Rollback
		sourceAccount.Deposit(amount)
		s.accountRepo.Update(sourceAccount)
		destinationAccount.Withdraw(amount)
		return err
	}

	// Record transaction
	transaction := models.NewTransfer(
		s.idGenerator.GenerateID(),
		sourceAccountID,
		destinationAccountID,
		amount,
		"Transfer",
	)

	return s.transactionRepo.Create(transaction)
}

// GetAccountTransactions retrieves all transactions for an account
func (s *DefaultBankingService) GetAccountTransactions(accountID string) ([]*models.Transaction, error) {
	return s.transactionRepo.GetByAccountID(accountID)
}
