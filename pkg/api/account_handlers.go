package api

import (
	"bankingsystem/pkg/models"
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// AccountRequest represents the request body for creating an account
type AccountRequest struct {
	CustomerID     string             `json:"customer_id" xml:"customer_id"`
	InitialBalance float64            `json:"initial_balance" xml:"initial_balance"`
	AccountType    models.AccountType `json:"account_type" xml:"account_type"`
}

// AccountResponse represents the response body for account operations
type AccountResponse struct {
	ID           string             `json:"id" xml:"id"`
	CustomerID   string             `json:"customer_id" xml:"customer_id"`
	Balance      float64            `json:"balance" xml:"balance"`
	AccountType  models.AccountType `json:"account_type" xml:"account_type"`
	CreatedAt    string             `json:"created_at" xml:"created_at"`
	LastActivity string             `json:"last_activity" xml:"last_activity"`
}

// AccountsResponse represents a list of accounts for response
type AccountsResponse struct {
	XMLName  xml.Name          `json:"-" xml:"accounts"`
	Accounts []AccountResponse `json:"accounts" xml:"account"`
}

// MoneyRequest represents a request for deposit, withdrawal, or transfer operations
type MoneyRequest struct {
	Amount               float64 `json:"amount" xml:"amount"`
	DestinationAccountID string  `json:"destination_account_id,omitempty" xml:"destination_account_id,omitempty"`
}

// handleGetAccount handles GET /accounts/{id}
func (s *Server) handleGetAccount(w http.ResponseWriter, r *http.Request, id string) {
	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}

	if account == nil {
		respondWithError(w, r, http.StatusNotFound, "Account not found")
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

	respond(w, r, http.StatusOK, response)
}

// handleCreateAccount handles POST /accounts
func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	var accountReq AccountRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := json.NewDecoder(r.Body).Decode(&accountReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := xml.NewDecoder(r.Body).Decode(&accountReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else {
		respondWithError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if accountReq.CustomerID == "" {
		respondWithError(w, r, http.StatusBadRequest, "Customer ID is required")
		return
	}

	if accountReq.InitialBalance < 0 {
		respondWithError(w, r, http.StatusBadRequest, "Initial balance cannot be negative")
		return
	}

	if accountReq.AccountType == "" {
		respondWithError(w, r, http.StatusBadRequest, "Account type is required")
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
			respondWithError(w, r, http.StatusNotFound, "Customer not found")
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to create account")
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

	respond(w, r, http.StatusCreated, response)
}

// handleDeleteAccount handles DELETE /accounts/{id}
func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request, id string) {
	err := s.bankingService.DeleteAccount(id)
	if err != nil {
		if err.Error() == "account not found" {
			respondWithError(w, r, http.StatusNotFound, "Account not found")
		} else if err.Error() == "cannot delete account with positive balance" {
			respondWithError(w, r, http.StatusConflict, "Cannot delete account with positive balance")
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to delete account")
		}
		return
	}

	// Return no content on successful delete
	w.WriteHeader(http.StatusNoContent)
}

// handleGetCustomerAccounts handles GET /customers/{id}/accounts
func (s *Server) handleGetCustomerAccounts(w http.ResponseWriter, r *http.Request, id string) {
	// First check if customer exists
	customer, err := s.bankingService.GetCustomer(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve customer")
		return
	}

	if customer == nil {
		respondWithError(w, r, http.StatusNotFound, "Customer not found")
		return
	}

	// Get accounts for customer
	accounts, err := s.bankingService.GetCustomerAccounts(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve accounts")
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

	respond(w, r, http.StatusOK, response)
}

// handleDeposit handles POST /accounts/{id}/deposit
func (s *Server) handleDeposit(w http.ResponseWriter, r *http.Request, id string) {
	contentType := r.Header.Get("Content-Type")
	var moneyReq MoneyRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := json.NewDecoder(r.Body).Decode(&moneyReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := xml.NewDecoder(r.Body).Decode(&moneyReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else {
		respondWithError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if moneyReq.Amount <= 0 {
		respondWithError(w, r, http.StatusBadRequest, "Amount must be positive")
		return
	}

	// Process deposit
	_, err := s.bankingService.Deposit(id, moneyReq.Amount, "Deposit from web API")
	if err != nil {
		if err.Error() == "account not found" {
			respondWithError(w, r, http.StatusNotFound, "Account not found")
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to process deposit")
		}
		return
	}

	// Get updated account
	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve updated account")
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

	respond(w, r, http.StatusOK, response)
}

// handleWithdraw handles POST /accounts/{id}/withdraw
func (s *Server) handleWithdraw(w http.ResponseWriter, r *http.Request, id string) {
	contentType := r.Header.Get("Content-Type")
	var moneyReq MoneyRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := json.NewDecoder(r.Body).Decode(&moneyReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := xml.NewDecoder(r.Body).Decode(&moneyReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else {
		respondWithError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Validate request
	if moneyReq.Amount <= 0 {
		respondWithError(w, r, http.StatusBadRequest, "Amount must be positive")
		return
	}

	// Process withdrawal
	_, err := s.bankingService.Withdraw(id, moneyReq.Amount, "Withdrawal from web API")
	if err != nil {
		if err.Error() == "account not found" {
			respondWithError(w, r, http.StatusNotFound, "Account not found")
		} else if err.Error() == "insufficient funds" {
			respondWithError(w, r, http.StatusBadRequest, "Insufficient funds")
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to process withdrawal")
		}
		return
	}

	// Get updated account
	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve updated account")
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

	respond(w, r, http.StatusOK, response)
}
