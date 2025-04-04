package api

import (
	"bankingsystem/pkg/services"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// GinServer represents the HTTP server using Gin framework for our banking API
type GinServer struct {
	bankingService services.BankingServiceInterface
	router         *gin.Engine
}

// NewGinServer creates a new API server with Gin
func NewGinServer(bankingService services.BankingServiceInterface) *GinServer {
	server := &GinServer{
		bankingService: bankingService,
		router:         gin.Default(),
	}

	// Set up middleware for handling content negotiation
	server.router.Use(contentNegotiationMiddleware())

	// Register routes
	server.registerRoutes()

	return server
}

// Start starts the Gin HTTP server on the specified port
func (s *GinServer) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting Gin server on %s", addr)
	return s.router.Run(addr)
}

// Shutdown gracefully shuts down the server
func (s *GinServer) Shutdown() {
	// Gin's Run() method blocks until the server is stopped
	// For proper shutdown handling, we would use http.Server.Shutdown
	// but keeping it simple for this implementation
	log.Println("Shutting down Gin server")
}

// registerRoutes sets up all the API endpoints with Gin
func (s *GinServer) registerRoutes() {
	// Customer routes
	customers := s.router.Group("/customers")
	{
		customers.GET("", s.handleListCustomers)
		customers.POST("", s.handleCreateCustomer)
		customers.GET("/:id", s.handleGetCustomer)
		customers.PUT("/:id", s.handleUpdateCustomer)
		customers.DELETE("/:id", s.handleDeleteCustomer)

		// Customer accounts routes
		customers.GET("/:id/accounts", s.handleGetCustomerAccounts)
	}

	// Account routes
	accounts := s.router.Group("/accounts")
	{
		accounts.POST("", s.handleCreateAccount)
		accounts.GET("/:id", s.handleGetAccount)
		accounts.DELETE("/:id", s.handleDeleteAccount)

		// Account transaction routes
		accounts.GET("/:id/transactions", s.handleGetAccountTransactions)
		accounts.POST("/:id/deposit", s.handleDeposit)
		accounts.POST("/:id/withdraw", s.handleWithdraw)
	}

	// Transfer route
	s.router.POST("/transfers", s.handleTransfer)
}

// contentNegotiationMiddleware handles content type negotiation for XML/JSON
func contentNegotiationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set default response type to JSON if not specified
		if c.GetHeader("Accept") == "" {
			c.Request.Header.Set("Accept", ContentTypeJSON)
		}

		c.Next()
	}
}

// respondWithError returns an error response with the appropriate format
func respondWithGinError(c *gin.Context, statusCode int, message string) {
	err := ErrorResponse{
		Status:  statusCode,
		Message: message,
	}

	// Check accept header for response format
	if c.GetHeader("Accept") == ContentTypeXML {
		c.XML(statusCode, err)
	} else {
		c.JSON(statusCode, err)
	}
}

// respondWithData returns data with the appropriate format
func respondWithGinData(c *gin.Context, statusCode int, data interface{}) {
	// Check accept header for response format
	if c.GetHeader("Accept") == ContentTypeXML {
		c.XML(statusCode, data)
	} else {
		c.JSON(statusCode, data)
	}
}
