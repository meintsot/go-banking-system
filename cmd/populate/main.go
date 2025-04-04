package main

import (
	"bankingsystem/pkg/database"
	"bankingsystem/pkg/models"
	"bankingsystem/pkg/utils"
	"log"
	"time"
)

func main() {
	log.Println("Populating Banking System database with test data...")

	// Initialize the database with GORM
	dbConn, err := database.NewGormDBConnection("banking.db")
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

	// Create test customers
	customers := []*models.Customer{
		{
			ID:        idGenerator.GenerateID(),
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "123-456-7890",
			Address:   "123 Main St, Anytown, USA",
		},
		{
			ID:        idGenerator.GenerateID(),
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane.smith@example.com",
			Phone:     "987-654-3210",
			Address:   "456 Oak Ave, Somewhere, USA",
		},
		{
			ID:        idGenerator.GenerateID(),
			FirstName: "Alice",
			LastName:  "Johnson",
			Email:     "alice.johnson@example.com",
			Phone:     "555-123-4567",
			Address:   "789 Pine Rd, Nowhere, USA",
		},
	}

	log.Println("Creating test customers...")
	for _, customer := range customers {
		log.Printf("Creating customer: %s %s", customer.FirstName, customer.LastName)
		if err := customerRepo.Create(customer); err != nil {
			log.Fatalf("Failed to create customer: %v", err)
		}
	}

	// Create accounts for each customer
	log.Println("Creating test accounts...")
	accounts := make([]*models.Account, 0)

	// Checking and Savings accounts for John
	johnCheckingID := idGenerator.GenerateID()
	johnSavingsID := idGenerator.GenerateID()

	now := time.Now()
	accounts = append(accounts,
		&models.Account{
			ID:           johnCheckingID,
			CustomerID:   customers[0].ID,
			Balance:      1000.00,
			AccountType:  models.Checking,
			CreatedAt:    now,
			LastActivity: now,
		},
		&models.Account{
			ID:           johnSavingsID,
			CustomerID:   customers[0].ID,
			Balance:      5000.00,
			AccountType:  models.Savings,
			CreatedAt:    now,
			LastActivity: now,
		},
	)

	// Checking account for Jane
	janeSavingsID := idGenerator.GenerateID()
	accounts = append(accounts,
		&models.Account{
			ID:           janeSavingsID,
			CustomerID:   customers[1].ID,
			Balance:      3000.00,
			AccountType:  models.Savings,
			CreatedAt:    now,
			LastActivity: now,
		},
	)

	// Business account for Alice
	aliceBusinessID := idGenerator.GenerateID()
	accounts = append(accounts,
		&models.Account{
			ID:           aliceBusinessID,
			CustomerID:   customers[2].ID,
			Balance:      10000.00,
			AccountType:  models.Business,
			CreatedAt:    now,
			LastActivity: now,
		},
	)

	for _, account := range accounts {
		log.Printf("Creating account: %s for customer %s", account.ID, account.CustomerID)
		if err := accountRepo.Create(account); err != nil {
			log.Fatalf("Failed to create account: %v", err)
		}
	}

	// Create some transactions
	log.Println("Creating test transactions...")
	transactions := []*models.Transaction{
		// Deposit to John's checking account
		{
			ID:              idGenerator.GenerateID(),
			AccountID:       johnCheckingID,
			Amount:          500.00,
			TransactionType: models.Deposit,
			Description:     "Initial deposit",
			Timestamp:       now.Add(-72 * time.Hour),
		},
		// Withdrawal from John's checking account
		{
			ID:              idGenerator.GenerateID(),
			AccountID:       johnCheckingID,
			Amount:          200.00,
			TransactionType: models.Withdrawal,
			Description:     "ATM withdrawal",
			Timestamp:       now.Add(-48 * time.Hour),
		},
		// Transfer from John's checking to savings
		{
			ID:                   idGenerator.GenerateID(),
			AccountID:            johnCheckingID,
			Amount:               300.00,
			TransactionType:      models.Transfer,
			Description:          "Transfer to savings",
			Timestamp:            now.Add(-24 * time.Hour),
			DestinationAccountID: johnSavingsID,
		},
		// Deposit to Jane's savings account
		{
			ID:              idGenerator.GenerateID(),
			AccountID:       janeSavingsID,
			Amount:          1000.00,
			TransactionType: models.Deposit,
			Description:     "Bonus deposit",
			Timestamp:       now.Add(-12 * time.Hour),
		},
		// Deposit to Alice's business account
		{
			ID:              idGenerator.GenerateID(),
			AccountID:       aliceBusinessID,
			Amount:          5000.00,
			TransactionType: models.Deposit,
			Description:     "Client payment",
			Timestamp:       now.Add(-6 * time.Hour),
		},
	}

	for _, transaction := range transactions {
		log.Printf("Creating transaction: %s for account %s", transaction.ID, transaction.AccountID)
		if err := transactionRepo.Create(transaction); err != nil {
			log.Fatalf("Failed to create transaction: %v", err)
		}
	}

	log.Println("Database successfully populated with test data!")

	// Verify the data by retrieving and displaying it
	log.Println("\nVerifying data:")

	// Get and display all customers
	allCustomers, err := customerRepo.GetAll()
	if err != nil {
		log.Fatalf("Failed to retrieve customers: %v", err)
	}
	log.Printf("Retrieved %d customers", len(allCustomers))
	for _, c := range allCustomers {
		log.Printf("Customer: %s %s (%s)", c.FirstName, c.LastName, c.Email)

		// Get and display accounts for this customer
		customerAccounts, err := accountRepo.GetByCustomerID(c.ID)
		if err != nil {
			log.Fatalf("Failed to retrieve accounts for customer %s: %v", c.ID, err)
		}

		for _, a := range customerAccounts {
			log.Printf("  Account: %s, Type: %s, Balance: $%.2f", a.ID, a.AccountType, a.Balance)

			// Get and display transactions for this account
			accountTransactions, err := transactionRepo.GetByAccountID(a.ID)
			if err != nil {
				log.Fatalf("Failed to retrieve transactions for account %s: %v", a.ID, err)
			}

			for _, t := range accountTransactions {
				if t.TransactionType == models.Transfer {
					log.Printf("    Transaction: %s, Type: %s, Amount: $%.2f, To: %s",
						t.ID, t.TransactionType, t.Amount, t.DestinationAccountID)
				} else {
					log.Printf("    Transaction: %s, Type: %s, Amount: $%.2f",
						t.ID, t.TransactionType, t.Amount)
				}
			}
		}
	}
}
