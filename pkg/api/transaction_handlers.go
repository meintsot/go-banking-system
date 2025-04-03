package api

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// TransactionResponse represents the response body for transaction operations
type TransactionResponse struct {
	ID                   string  `json:"id" xml:"id"`
	AccountID            string  `json:"account_id" xml:"account_id"`
	Amount               float64 `json:"amount" xml:"amount"`
	TransactionType      string  `json:"transaction_type" xml:"transaction_type"`
	Description          string  `json:"description" xml:"description"`
	Timestamp            string  `json:"timestamp" xml:"timestamp"`
	DestinationAccountID string  `json:"destination_account_id,omitempty" xml:"destination_account_id,omitempty"`
}

// TransactionsResponse represents a list of transactions for response
type TransactionsResponse struct {
	XMLName      xml.Name              `json:"-" xml:"transactions"`
	Transactions []TransactionResponse `json:"transactions" xml:"transaction"`
}

// handleGetAccountTransactions handles GET /accounts/{id}/transactions
func (s *Server) handleGetAccountTransactions(w http.ResponseWriter, r *http.Request, id string) {
	// First check if account exists
	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}

	if account == nil {
		respondWithError(w, r, http.StatusNotFound, "Account not found")
		return
	}

	// Get transactions for account
	transactions, err := s.bankingService.GetAccountTransactions(id)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve transactions")
		return
	}

	// Map to response type
	response := TransactionsResponse{
		Transactions: make([]TransactionResponse, len(transactions)),
	}

	for i, transaction := range transactions {
		response.Transactions[i] = TransactionResponse{
			ID:                   transaction.ID,
			AccountID:            transaction.AccountID,
			Amount:               transaction.Amount,
			TransactionType:      string(transaction.TransactionType),
			Description:          transaction.Description,
			Timestamp:            transaction.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			DestinationAccountID: transaction.DestinationAccountID,
		}
	}

	respond(w, r, http.StatusOK, response)
}

// handleTransfer handles POST /transfers
func (s *Server) handleTransfer(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	var transferReq MoneyRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := json.NewDecoder(r.Body).Decode(&transferReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := xml.NewDecoder(r.Body).Decode(&transferReq); err != nil {
			respondWithError(w, r, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else {
		respondWithError(w, r, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Get source and destination account IDs from URL and request body
	sourceAccountID := r.URL.Query().Get("from")
	destinationAccountID := transferReq.DestinationAccountID

	// Validate request
	if sourceAccountID == "" {
		respondWithError(w, r, http.StatusBadRequest, "Source account ID is required (use ?from=source_account_id)")
		return
	}

	if destinationAccountID == "" {
		respondWithError(w, r, http.StatusBadRequest, "Destination account ID is required")
		return
	}

	if transferReq.Amount <= 0 {
		respondWithError(w, r, http.StatusBadRequest, "Transfer amount must be positive")
		return
	}

	// Process transfer
	transaction, err := s.bankingService.Transfer(sourceAccountID, destinationAccountID, transferReq.Amount, "Transfer from web API")
	if err != nil {
		if err.Error() == "source account not found" {
			respondWithError(w, r, http.StatusNotFound, "Source account not found")
		} else if err.Error() == "destination account not found" {
			respondWithError(w, r, http.StatusNotFound, "Destination account not found")
		} else if err.Error() == "insufficient funds" {
			respondWithError(w, r, http.StatusBadRequest, "Insufficient funds")
		} else {
			respondWithError(w, r, http.StatusInternalServerError, "Failed to process transfer")
		}
		return
	}

	// Get updated source account to return in the response
	sourceAccount, err := s.bankingService.GetAccount(sourceAccountID)
	if err != nil {
		respondWithError(w, r, http.StatusInternalServerError, "Failed to retrieve updated source account")
		return
	}

	// Create response with both transaction and account information
	type TransferResponse struct {
		Transaction TransactionResponse `json:"transaction" xml:"transaction"`
		Account     AccountResponse     `json:"account" xml:"account"`
	}

	accountResp := AccountResponse{
		ID:           sourceAccount.ID,
		CustomerID:   sourceAccount.CustomerID,
		Balance:      sourceAccount.Balance,
		AccountType:  sourceAccount.AccountType,
		CreatedAt:    sourceAccount.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastActivity: sourceAccount.LastActivity.Format("2006-01-02T15:04:05Z07:00"),
	}

	transactionResp := TransactionResponse{
		ID:                   transaction.ID,
		AccountID:            transaction.AccountID,
		Amount:               transaction.Amount,
		TransactionType:      string(transaction.TransactionType),
		Description:          transaction.Description,
		Timestamp:            transaction.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		DestinationAccountID: transaction.DestinationAccountID,
	}

	response := TransferResponse{
		Transaction: transactionResp,
		Account:     accountResp,
	}

	respond(w, r, http.StatusOK, response)
}
