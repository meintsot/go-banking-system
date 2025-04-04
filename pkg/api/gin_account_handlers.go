package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleGetAccount handles GET /accounts/:id
func (s *GinServer) handleGetAccount(c *gin.Context) {
	id := c.Param("id")

	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}

	if account == nil {
		respondWithGinError(c, http.StatusNotFound, "Account not found")
		return
	}

	response := AccountResponse{
		ID:           account.ID,
		CustomerID:   account.CustomerID,
		Balance:      account.Balance,
		AccountType:  account.AccountType,
		CreatedAt:    account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastActivity: account.LastActivity.Format("2006-01-02T15:04:05Z07:00"),
	}

	respondWithGinData(c, http.StatusOK, response)
}

// handleCreateAccount handles POST /accounts
func (s *GinServer) handleCreateAccount(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	var accountReq AccountRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := c.ShouldBindJSON(&accountReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := c.ShouldBindXML(&accountReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid XML request body")
			return
		}
	} else {
		respondWithGinError(c, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if accountReq.CustomerID == "" {
		respondWithGinError(c, http.StatusBadRequest, "Customer ID is required")
		return
	}

	if accountReq.InitialBalance < 0 {
		respondWithGinError(c, http.StatusBadRequest, "Initial balance cannot be negative")
		return
	}

	if accountReq.AccountType == "" {
		respondWithGinError(c, http.StatusBadRequest, "Account type is required")
		return
	}

	// Create account
	account, err := s.bankingService.CreateAccount(
		accountReq.CustomerID,
		string(accountReq.AccountType),
		accountReq.InitialBalance,
	)
	if err != nil {
		if err.Error() == "customer not found" {
			respondWithGinError(c, http.StatusNotFound, "Customer not found")
		} else {
			respondWithGinError(c, http.StatusInternalServerError, "Failed to create account")
		}
		return
	}

	// Return created account
	response := AccountResponse{
		ID:           account.ID,
		CustomerID:   account.CustomerID,
		Balance:      account.Balance,
		AccountType:  account.AccountType,
		CreatedAt:    account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastActivity: account.LastActivity.Format("2006-01-02T15:04:05Z07:00"),
	}

	respondWithGinData(c, http.StatusCreated, response)
}

// handleDeleteAccount handles DELETE /accounts/:id
func (s *GinServer) handleDeleteAccount(c *gin.Context) {
	id := c.Param("id")

	err := s.bankingService.DeleteAccount(id)
	if err != nil {
		if err.Error() == "account not found" {
			respondWithGinError(c, http.StatusNotFound, "Account not found")
		} else if err.Error() == "cannot delete account with positive balance" {
			respondWithGinError(c, http.StatusConflict, "Cannot delete account with positive balance")
		} else {
			respondWithGinError(c, http.StatusInternalServerError, "Failed to delete account")
		}
		return
	}

	// Return no content on successful delete
	c.Status(http.StatusNoContent)
}

// handleGetCustomerAccounts handles GET /customers/:id/accounts
func (s *GinServer) handleGetCustomerAccounts(c *gin.Context) {
	id := c.Param("id")

	// First check if customer exists
	customer, err := s.bankingService.GetCustomer(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		respondWithGinError(c, http.StatusNotFound, "Customer not found")
		return
	}

	// Get accounts for customer
	accounts, err := s.bankingService.GetCustomerAccounts(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve accounts")
		return
	}

	// Map to response type
	response := AccountsResponse{
		Accounts: make([]AccountResponse, len(accounts)),
	}

	for i, account := range accounts {
		response.Accounts[i] = AccountResponse{
			ID:           account.ID,
			CustomerID:   account.CustomerID,
			Balance:      account.Balance,
			AccountType:  account.AccountType,
			CreatedAt:    account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastActivity: account.LastActivity.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	respondWithGinData(c, http.StatusOK, response)
}

// handleDeposit handles POST /accounts/:id/deposit
func (s *GinServer) handleDeposit(c *gin.Context) {
	id := c.Param("id")
	contentType := c.GetHeader("Content-Type")
	var moneyReq MoneyRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := c.ShouldBindJSON(&moneyReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := c.ShouldBindXML(&moneyReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid XML request body")
			return
		}
	} else {
		respondWithGinError(c, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if moneyReq.Amount <= 0 {
		respondWithGinError(c, http.StatusBadRequest, "Amount must be positive")
		return
	}

	// Process deposit
	_, err := s.bankingService.Deposit(id, moneyReq.Amount, "Deposit from web API")
	if err != nil {
		if err.Error() == "account not found" {
			respondWithGinError(c, http.StatusNotFound, "Account not found")
		} else {
			respondWithGinError(c, http.StatusInternalServerError, "Failed to process deposit")
		}
		return
	}

	// Get updated account
	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve updated account")
		return
	}

	// Return updated account
	response := AccountResponse{
		ID:           account.ID,
		CustomerID:   account.CustomerID,
		Balance:      account.Balance,
		AccountType:  account.AccountType,
		CreatedAt:    account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastActivity: account.LastActivity.Format("2006-01-02T15:04:05Z07:00"),
	}

	respondWithGinData(c, http.StatusOK, response)
}

// handleWithdraw handles POST /accounts/:id/withdraw
func (s *GinServer) handleWithdraw(c *gin.Context) {
	id := c.Param("id")
	contentType := c.GetHeader("Content-Type")
	var moneyReq MoneyRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := c.ShouldBindJSON(&moneyReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := c.ShouldBindXML(&moneyReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid XML request body")
			return
		}
	} else {
		respondWithGinError(c, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if moneyReq.Amount <= 0 {
		respondWithGinError(c, http.StatusBadRequest, "Amount must be positive")
		return
	}

	// Process withdrawal
	_, err := s.bankingService.Withdraw(id, moneyReq.Amount, "Withdrawal from web API")
	if err != nil {
		if err.Error() == "account not found" {
			respondWithGinError(c, http.StatusNotFound, "Account not found")
		} else if err.Error() == "insufficient funds" {
			respondWithGinError(c, http.StatusBadRequest, "Insufficient funds")
		} else {
			respondWithGinError(c, http.StatusInternalServerError, "Failed to process withdrawal")
		}
		return
	}

	// Get updated account
	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve updated account")
		return
	}

	// Return updated account
	response := AccountResponse{
		ID:           account.ID,
		CustomerID:   account.CustomerID,
		Balance:      account.Balance,
		AccountType:  account.AccountType,
		CreatedAt:    account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastActivity: account.LastActivity.Format("2006-01-02T15:04:05Z07:00"),
	}

	respondWithGinData(c, http.StatusOK, response)
}
