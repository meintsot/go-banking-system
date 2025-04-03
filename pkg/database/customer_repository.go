package database

import (
	"bankingsystem/pkg/models"
	"database/sql"
	"errors"
	"fmt"
)

// SQLiteCustomerRepository implements customer repository using SQLite
type SQLiteCustomerRepository struct {
	db *sql.DB
}

// NewSQLiteCustomerRepository creates a new SQLite customer repository
func NewSQLiteCustomerRepository(conn *DBConnection) *SQLiteCustomerRepository {
	return &SQLiteCustomerRepository{
		db: conn.GetDB(),
	}
}

// Create inserts a new customer
func (r *SQLiteCustomerRepository) Create(customer *models.Customer) error {
	query := `INSERT INTO customers (id, first_name, last_name, email, phone, address) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		customer.ID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.Address)

	if err != nil {
		return fmt.Errorf("error creating customer: %v", err)
	}

	return nil
}

// GetByID retrieves a customer by ID
func (r *SQLiteCustomerRepository) GetByID(id string) (*models.Customer, error) {
	query := `SELECT id, first_name, last_name, email, phone, address FROM customers WHERE id = ?`

	var customer models.Customer
	err := r.db.QueryRow(query, id).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("error getting customer: %v", err)
	}

	return &customer, nil
}

// GetAll retrieves all customers
func (r *SQLiteCustomerRepository) GetAll() ([]*models.Customer, error) {
	query := `SELECT id, first_name, last_name, email, phone, address FROM customers`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting customers: %v", err)
	}
	defer rows.Close()

	customers := []*models.Customer{}

	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email,
			&customer.Phone,
			&customer.Address,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning customer row: %v", err)
		}

		customers = append(customers, &customer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customer rows: %v", err)
	}

	return customers, nil
}

// List returns all customers (alias for GetAll to satisfy interface)
func (r *SQLiteCustomerRepository) List() ([]*models.Customer, error) {
	return r.GetAll()
}

// Update updates a customer
func (r *SQLiteCustomerRepository) Update(customer *models.Customer) error {
	query := `UPDATE customers 
			  SET first_name = ?, last_name = ?, email = ?, phone = ?, address = ? 
			  WHERE id = ?`

	result, err := r.db.Exec(query,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.Address,
		customer.ID)

	if err != nil {
		return fmt.Errorf("error updating customer: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return errors.New("customer not found")
	}

	return nil
}

// Delete deletes a customer
func (r *SQLiteCustomerRepository) Delete(id string) error {
	// Check if customer has any accounts
	query := `SELECT COUNT(*) FROM accounts WHERE customer_id = ?`
	var count int
	err := r.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking customer accounts: %v", err)
	}

	if count > 0 {
		return errors.New("cannot delete customer with active accounts")
	}

	// Delete the customer
	query = `DELETE FROM customers WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting customer: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return errors.New("customer not found")
	}

	return nil
}

// GetByEmail retrieves a customer by email
func (r *SQLiteCustomerRepository) GetByEmail(email string) (*models.Customer, error) {
	query := `SELECT id, first_name, last_name, email, phone, address FROM customers WHERE email = ?`

	var customer models.Customer
	err := r.db.QueryRow(query, email).Scan(
		&customer.ID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.Address,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("error getting customer by email: %v", err)
	}

	return &customer, nil
}
