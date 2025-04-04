package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleGetAccountTransactions handles GET /accounts/:id/transactions
func (s *GinServer) handleGetAccountTransactions(c *gin.Context) {
	id := c.Param("id")

	// First check if account exists
	account, err := s.bankingService.GetAccount(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}

	if account == nil {
		respondWithGinError(c, http.StatusNotFound, "Account not found")
		return
	}

	// Get transactions for account
	transactions, err := s.bankingService.GetAccountTransactions(id)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve transactions")
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

	respondWithGinData(c, http.StatusOK, response)
}

// handleTransfer handles POST /transfers
func (s *GinServer) handleTransfer(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	var transferReq MoneyRequest

	// Parse request body based on Content-Type
	if contentType == ContentTypeJSON || contentType == ContentTypeJSON+"; charset=utf-8" {
		if err := c.ShouldBindJSON(&transferReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
	} else if contentType == ContentTypeXML || contentType == ContentTypeXML+"; charset=utf-8" {
		if err := c.ShouldBindXML(&transferReq); err != nil {
			respondWithGinError(c, http.StatusBadRequest, "Invalid XML request body")
			return
		}
	} else {
		respondWithGinError(c, http.StatusUnsupportedMediaType, "Content-Type must be application/json or application/xml")
		return
	}

	// Get source account ID from query parameter
	sourceAccountID := c.Query("from")
	destinationAccountID := transferReq.DestinationAccountID

	// Validate request
	if sourceAccountID == "" {
		respondWithGinError(c, http.StatusBadRequest, "Source account ID is required (use ?from=source_account_id)")
		return
	}

	if destinationAccountID == "" {
		respondWithGinError(c, http.StatusBadRequest, "Destination account ID is required")
		return
	}

	if transferReq.Amount <= 0 {
		respondWithGinError(c, http.StatusBadRequest, "Transfer amount must be positive")
		return
	}

	// Process transfer
	transaction, err := s.bankingService.Transfer(sourceAccountID, destinationAccountID, transferReq.Amount, "Transfer from web API")
	if err != nil {
		if err.Error() == "source account not found" || err.Error() == "account with ID "+sourceAccountID+" not found" {
			respondWithGinError(c, http.StatusNotFound, "Source account not found")
		} else if err.Error() == "destination account not found" || err.Error() == "account with ID "+destinationAccountID+" not found" {
			respondWithGinError(c, http.StatusNotFound, "Destination account not found")
		} else if err.Error() == "insufficient funds" || err.Error() == "insufficient funds in source account" {
			respondWithGinError(c, http.StatusBadRequest, "Insufficient funds")
		} else {
			respondWithGinError(c, http.StatusInternalServerError, "Failed to process transfer")
		}
		return
	}

	// Get updated source account to return in the response
	sourceAccount, err := s.bankingService.GetAccount(sourceAccountID)
	if err != nil {
		respondWithGinError(c, http.StatusInternalServerError, "Failed to retrieve updated source account")
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

	respondWithGinData(c, http.StatusOK, response)
}
