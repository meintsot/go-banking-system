package models

// Customer represents a bank customer
type Customer struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
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
