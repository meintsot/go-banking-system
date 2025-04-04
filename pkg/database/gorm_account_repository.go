package database

import (
	"bankingsystem/pkg/models"
	"errors"

	"gorm.io/gorm"
)

// GormAccountRepository implements account repository using GORM
type GormAccountRepository struct {
	db *gorm.DB
}

// NewGormAccountRepository creates a new GORM account repository
func NewGormAccountRepository(conn *GormDBConnection) *GormAccountRepository {
	return &GormAccountRepository{
		db: conn.GetDB(),
	}
}

// Create inserts a new account
func (r *GormAccountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

// GetByID retrieves an account by ID
func (r *GormAccountRepository) GetByID(id string) (*models.Account, error) {
	var account models.Account
	result := r.db.First(&account, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, result.Error
	}

	return &account, nil
}

// GetByCustomerID retrieves all accounts for a specific customer
func (r *GormAccountRepository) GetByCustomerID(customerID string) ([]*models.Account, error) {
	var accounts []*models.Account
	result := r.db.Where("customer_id = ?", customerID).Find(&accounts)

	if result.Error != nil {
		return nil, result.Error
	}

	return accounts, nil
}

// Update updates an account
func (r *GormAccountRepository) Update(account *models.Account) error {
	result := r.db.Save(account)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("account not found")
	}

	return nil
}

// Delete deletes an account
func (r *GormAccountRepository) Delete(id string) error {
	// GORM will use the BeforeDelete hook we defined in the model
	result := r.db.Delete(&models.Account{}, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, models.ErrorAccountHasBalance) {
			return errors.New("cannot delete account with positive balance")
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("account not found")
	}

	return nil
}
