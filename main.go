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

	// Check if the database has data before starting
	hasExistingData := false
	if _, err := os.Stat(dbPath); err == nil {
		// Check if the database has data
		tempConn, tempErr := database.NewGormDBConnection(dbPath)
		if tempErr == nil {
			defer tempConn.Close()

			var customerCount int64
			tempConn.GetDB().Table("customers").Count(&customerCount)

			// If we have data, use the existing database
			if customerCount > 0 {
				log.Printf("Found existing database with %d customers, using it without reinitializing", customerCount)
				hasExistingData = true
			} else {
				// No data, so back up the empty database
				backupPath := "banking_backup_" + time.Now().Format("20060102_150405") + ".db"
				log.Printf("Creating backup of empty database to %s", backupPath)
				tempConn.Close() // Close before renaming
				if err := utils.CopyFile(dbPath, backupPath); err != nil {
					log.Printf("Warning: Could not create backup of database: %v", err)
				}
			}
		} else {
			// Couldn't open the database, create a backup just in case
			backupPath := "banking_backup_" + time.Now().Format("20060102_150405") + ".db"
			log.Printf("Error opening database, creating backup to %s: %v", backupPath, tempErr)
			if err := utils.CopyFile(dbPath, backupPath); err != nil {
				log.Printf("Warning: Could not create backup of database: %v", err)
			}
		}
	}

	// Initialize the database with GORM
	dbConn, err := database.NewGormDBConnection(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	// Initialize database schema if needed (this won't overwrite data)
	if err := dbConn.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// If we have existing data, log it for verification
	if hasExistingData {
		var counts struct {
			Customers    int64
			Accounts     int64
			Transactions int64
		}

		dbConn.GetDB().Table("customers").Count(&counts.Customers)
		dbConn.GetDB().Table("accounts").Count(&counts.Accounts)
		dbConn.GetDB().Table("transactions").Count(&counts.Transactions)

		log.Printf("Database contains: %d customers, %d accounts, %d transactions",
			counts.Customers, counts.Accounts, counts.Transactions)
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

	// Start server in a goroutine - use port 8081 to avoid conflicts
	go func() {
		port := 8081
		log.Printf("Server starting on port %d...", port)
		if err := server.Start(port); err != nil {
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
