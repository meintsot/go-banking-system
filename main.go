package main

import (
	"bankingsystem/pkg/api"
	"bankingsystem/pkg/database"
	"bankingsystem/pkg/services"
	"bankingsystem/pkg/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting Banking System...")

	// Initialize the database
	dbConn, err := database.NewDBConnection("banking.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	// Initialize database schema if needed
	if err := dbConn.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Create ID generator
	idGenerator := utils.NewUUIDGenerator()

	// Initialize repositories
	customerRepo := database.NewSQLiteCustomerRepository(dbConn)
	accountRepo := database.NewSQLiteAccountRepository(dbConn)
	transactionRepo := database.NewSQLiteTransactionRepository(dbConn)

	// Initialize service
	bankingService := services.NewBankingService(customerRepo, accountRepo, transactionRepo, idGenerator)

	// Initialize and start server
	server := api.NewServer(bankingService)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port 8080...")
		if err := server.Start(8080); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	// Since there's no Shutdown method, we'll just log that the server has stopped
	log.Println("Server exited properly")
}
