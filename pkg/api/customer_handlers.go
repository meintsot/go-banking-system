package api

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// CustomerRequest represents the request body for customer operations
type CustomerRequest struct {
	FirstName string `json:"first_name" xml:"first_name"`
	LastName  string `json:"last_name" xml:"last_name"`
	Email     string `json:"email" xml:"email"`
	Phone     string `json:"phone" xml:"phone"`
	Address   string `json:"address" xml:"address"`
}

// CustomerResponse represents the response body for customer operations
type CustomerResponse struct {
	ID        string `json:"id" xml:"id"`
	FirstName string `json:"first_name" xml:"first_name"`
	LastName  string `json:"last_name" xml:"last_name"`
	Email     string `json:"email" xml:"email"`
	Phone     string `json:"phone" xml:"phone"`
	Address   string `json:"address" xml:"address"`
}

// CustomersResponse represents a list of customers for response
type CustomersResponse struct {
	XMLName   xml.Name           `json:"-" xml:"customers"`
	Customers []CustomerResponse `json:"customers" xml:"customer"`
}

// handleListCustomers handles GET /customers
func (s *Server) handleListCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := s.bankingService.ListCustomers()
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve customers")
		return
	}

	// Map to response type
	response := CustomersResponse{
		Customers: make([]CustomerResponse, len(customers)),
	}

	for i, customer := range customers {
		response.Customers[i] = CustomerResponse{
			ID:        customer.ID,
			FirstName: customer.FirstName,
			LastName:  customer.LastName,
			Email:     customer.Email,
			Phone:     customer.Phone,
			Address:   customer.Address,
		}
	}

	respond(w, r, http.StatusOK, response)
}

// handleGetCustomer handles GET /customers/{id}
func (s *Server) handleGetCustomer(w http.ResponseWriter, r *http.Request, id string) {
	customer, err := s.bankingService.GetCustomer(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		respondWithError(w, r, http.StatusNotFound, "Customer not found")
		return
	}

	response := CustomerResponse{
		ID:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		Address:   customer.Address,
	}

	respond(w, r, http.StatusOK, response)
}

// handleCreateCustomer handles POST /customers
func (s *Server) handleCreateCustomer(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	var customerReq CustomerRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := json.NewDecoder(r.Body).Decode(&customerReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := xml.NewDecoder(r.Body).Decode(&customerReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else {
		respondWithError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if customerReq.FirstName == "" || customerReq.LastName == "" || customerReq.Email == "" {
		respondWithError(w, r, http.StatusBadRequest, "First name, last name, and email are required")
		return
	}

	// Create customer
	customer, err := s.bankingService.CreateCustomer(
		customerReq.FirstName,
		customerReq.LastName,
		customerReq.Email,
		customerReq.Phone,
		customerReq.Address,
	)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to create customer")
		return
	}

	// Return response
	response := CustomerResponse{
		ID:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		Address:   customer.Address,
	}

	respond(w, r, http.StatusCreated, response)
}

// handleUpdateCustomer handles PUT /customers/{id}
func (s *Server) handleUpdateCustomer(w http.ResponseWriter, r *http.Request, id string) {
	// Check if customer exists
	customer, err := s.bankingService.GetCustomer(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		respondWithError(w, r, http.StatusNotFound, "Customer not found")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var customerReq CustomerRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := json.NewDecoder(r.Body).Decode(&customerReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := xml.NewDecoder(r.Body).Decode(&customerReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else {
		respondWithError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Update customer fields
	if customerReq.FirstName != "" {
		customer.FirstName = customerReq.FirstName
	}
	if customerReq.LastName != "" {
		customer.LastName = customerReq.LastName
	}
	if customerReq.Email != "" {
		customer.Email = customerReq.Email
	}
	if customerReq.Phone != "" {
		customer.Phone = customerReq.Phone
	}
	if customerReq.Address != "" {
		customer.Address = customerReq.Address
	}

	// Save changes
	err = s.bankingService.UpdateCustomer(customer)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to update customer")
		return
	}

	// Return updated customer
	response := CustomerResponse{
		ID:        customer.ID,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		Email:     customer.Email,
		Phone:     customer.Phone,
		Address:   customer.Address,
	}

	respond(w, r, http.StatusOK, response)
}

// handleDeleteCustomer handles DELETE /customers/{id}
func (s *Server) handleDeleteCustomer(w http.ResponseWriter, r *http.Request, id string) {
	err := s.bankingService.DeleteCustomer(id)
	if err != nil {
		// Check if error is "customer not found" or "cannot delete customer with active accounts"
		if err.Error() == "customer not found" {
			respondWithError(w, r, http.StatusNotFound, "Customer not found")
		} else if err.Error() == "cannot delete customer with active accounts" {
			respondWithError(w, r, http.StatusConflict, "Cannot delete customer with active accounts")
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to delete customer")
		}
		return
	}

	// Return no content on successful delete
	w.WriteHeader(http.StatusNoContent)
}
