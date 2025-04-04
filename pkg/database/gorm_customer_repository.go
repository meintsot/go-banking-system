package database

import (
	"bankingsystem/pkg/models"
	"errors"

	"gorm.io/gorm"
)

// GormCustomerRepository implements customer repository using GORM
type GormCustomerRepository struct {
	db *gorm.DB
}

// NewGormCustomerRepository creates a new GORM customer repository
func NewGormCustomerRepository(conn *GormDBConnection) *GormCustomerRepository {
	return &GormCustomerRepository{
		db: conn.GetDB(),
	}
}

// Create inserts a new customer
func (r *GormCustomerRepository) Create(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

// GetByID retrieves a customer by ID
func (r *GormCustomerRepository) GetByID(id string) (*models.Customer, error) {
	var customer models.Customer
	result := r.db.First(&customer, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, result.Error
	}

	return &customer, nil
}

// GetAll retrieves all customers
func (r *GormCustomerRepository) GetAll() ([]*models.Customer, error) {
	var customers []*models.Customer
	result := r.db.Find(&customers)

	if result.Error != nil {
		return nil, result.Error
	}

	return customers, nil
}

// List returns all customers (alias for GetAll to satisfy interface)
func (r *GormCustomerRepository) List() ([]*models.Customer, error) {
	return r.GetAll()
}

// Update updates a customer
func (r *GormCustomerRepository) Update(customer *models.Customer) error {
	result := r.db.Save(customer)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("customer not found")
	}

	return nil
}

// Delete deletes a customer
func (r *GormCustomerRepository) Delete(id string) error {
	// GORM will use the BeforeDelete hook we defined in the model
	result := r.db.Delete(&models.Customer{}, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, models.ErrorCustomerHasAccounts) {
			return errors.New("cannot delete customer with active accounts")
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("customer not found")
	}

	return nil
}

// GetByEmail retrieves a customer by email
func (r *GormCustomerRepository) GetByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	result := r.db.First(&customer, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, result.Error
	}

	return &customer, nil
}
