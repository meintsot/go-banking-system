package database

import (
	"bankingsystem/pkg/models"
	"errors"

	"gorm.io/gorm"
)

// GormTransactionRepository implements transaction repository using GORM
type GormTransactionRepository struct {
	db *gorm.DB
}

// NewGormTransactionRepository creates a new GORM transaction repository
func NewGormTransactionRepository(conn *GormDBConnection) *GormTransactionRepository {
	return &GormTransactionRepository{
		db: conn.GetDB(),
	}
}

// Create inserts a new transaction
func (r *GormTransactionRepository) Create(transaction *models.Transaction) error {
	return r.db.Create(transaction).Error
}

// GetByID retrieves a transaction by ID
func (r *GormTransactionRepository) GetByID(id string) (*models.Transaction, error) {
	var transaction models.Transaction
	result := r.db.First(&transaction, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, result.Error
	}

	return &transaction, nil
}

// GetByAccountID retrieves all transactions for a specific account
func (r *GormTransactionRepository) GetByAccountID(accountID string) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	result := r.db.Where("account_id = ?", accountID).
		Order("timestamp desc").
		Find(&transactions)

	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// GetTransfersByAccountID retrieves transfers (both incoming and outgoing) for an account
func (r *GormTransactionRepository) GetTransfersByAccountID(accountID string) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	// Get transactions where the account is either the source or destination of a transfer
	result := r.db.Where(
		"(account_id = ? OR destination_account_id = ?) AND transaction_type = ?",
		accountID, accountID, models.Transfer).
		Order("timestamp desc").
		Find(&transactions)

	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// CreateInTransaction creates a transaction within a database transaction
func (r *GormTransactionRepository) CreateInTransaction(tx *gorm.DB, transaction *models.Transaction) error {
	return tx.Create(transaction).Error
}
