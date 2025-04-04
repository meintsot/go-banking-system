package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleListCustomers handles GET /customers
func (s *GinServer) handleListCustomers(c *gin.Context) {
	customers, err := s.bankingService.ListCustomers()
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve customers")
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

	respondWithGinData(c, http.StatusOK, response)
}

// handleGetCustomer handles GET /customers/:id
func (s *GinServer) handleGetCustomer(c *gin.Context) {
	id := c.Param("id")

	customer, err := s.bankingService.GetCustomer(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		respondWithGinError(c, http.StatusNotFound, "Customer not found")
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

	respondWithGinData(c, http.StatusOK, response)
}

// handleCreateCustomer handles POST /customers
func (s *GinServer) handleCreateCustomer(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	var customerReq CustomerRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := c.ShouldBindJSON(&customerReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := c.ShouldBindXML(&customerReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid XML request body")
			return
		}
	} else {
		respondWithGinError(c, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if customerReq.FirstName == "" || customerReq.LastName == "" || customerReq.Email == "" {
		respondWithGinError(c, http.StatusBadRequest, "First name, last name, and email are required")
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
		respondWithGinError(c, http.StatusInternalServerError, "Failed to create customer")
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

	respondWithGinData(c, http.StatusCreated, response)
}

// handleUpdateCustomer handles PUT /customers/:id
func (s *GinServer) handleUpdateCustomer(c *gin.Context) {
	id := c.Param("id")

	// Check if customer exists
	customer, err := s.bankingService.GetCustomer(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		respondWithGinError(c, http.StatusNotFound, "Customer not found")
		return
	}

	contentType := c.GetHeader("Content-Type")
	var customerReq CustomerRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := c.ShouldBindJSON(&customerReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := c.ShouldBindXML(&customerReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid XML request body")
			return
		}
	} else {
		respondWithGinError(c, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
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
		respondWithGinError(c, http.StatusInternalServerError, "Failed to update customer")
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

	respondWithGinData(c, http.StatusOK, response)
}

// handleDeleteCustomer handles DELETE /customers/:id
func (s *GinServer) handleDeleteCustomer(c *gin.Context) {
	id := c.Param("id")

	err := s.bankingService.DeleteCustomer(id)
	if err != nil {
		// Check if error is "customer not found" or "cannot delete customer with active accounts"
		if err.Error() == "customer not found" {
			respondWithGinError(c, http.StatusNotFound, "Customer not found")
		} else if err.Error() == "cannot delete customer with active accounts" {
			respondWithGinError(c, http.StatusConflict, "Cannot delete customer with active accounts")
		} else {
			respondWithGinError(c, http.StatusInternalServerError, "Failed to delete customer")
		}
		return
	}

	// Return no content on successful delete
	c.Status(http.StatusNoContent)
}
