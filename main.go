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
	"time"
)

func main() {
	log.Println("Starting Banking System with Gin and GORM...")

	dbPath := "banking.db"

	// For a clean migration from standard SQL to GORM, rename the old database
	if _, err := os.Stat(dbPath); err == nil {
		// File exists, back it up
		backupPath := "banking_backup_" + time.Now().Format("20060102_150405") + ".db"
		log.Printf("Creating backup of existing database to %s", backupPath)
		if err := os.Rename(dbPath, backupPath); err != nil {
			log.Printf("Warning: Could not rename existing database: %v", err)
		}
	}

	// Initialize the database with GORM
	dbConn, err := database.NewGormDBConnection(dbPath)
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

	// Initialize repositories with GORM
	customerRepo := database.NewGormCustomerRepository(dbConn)
	accountRepo := database.NewGormAccountRepository(dbConn)
	transactionRepo := database.NewGormTransactionRepository(dbConn)

	// Initialize service
	bankingService := services.NewBankingService(customerRepo, accountRepo, transactionRepo, idGenerator)

	// Initialize and start Gin server
	server := api.NewGinServer(bankingService)

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
	server.Shutdown()
	log.Println("Server exited properly")
}
