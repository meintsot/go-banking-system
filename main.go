package main

import (
	"bankingsystem/pkg/adapters"
	"bankingsystem/pkg/models"
	"bankingsystem/pkg/services"
	"bankingsystem/pkg/utils"
	"fmt"
)

func main() {
	// Initialize repositories
	customerRepo := adapters.NewInMemoryCustomerRepository()
	accountRepo := adapters.NewInMemoryAccountRepository()
	transactionRepo := adapters.NewInMemoryTransactionRepository()

	// Initialize ID generator
	idGenerator := utils.NewUUIDGenerator()

	// Initialize banking service
	bankingService := services.NewBankingService(
		customerRepo,
		accountRepo,
		transactionRepo,
		idGenerator,
	)

	// Initialize adapters for external systems
	notificationAdapter := adapters.NewNotificationAdapter(true) // Start with email notifications
	reportingSystem := adapters.NewMockExternalReportingSystem()
	reportingAdapter := adapters.NewReportingAdapter(reportingSystem)

	// Simulate banking operations
	fmt.Println("=== Banking System Simulation ===")

	// Create a customer
	customer, err := bankingService.CreateCustomer(
		"John",
		"Doe",
		"john.doe@example.com",
		"555-123-4567",
		"123 Main St",
	)
	if err != nil {
		fmt.Printf("Error creating customer: %v\n", err)
		return
	}

	fmt.Printf("Created customer: %s %s (ID: %s)\n", customer.FirstName, customer.LastName, customer.ID)

	// Send welcome notification
	notificationAdapter.SendNotification(customer.ID, "Welcome to our bank!")

	// Create a checking account
	checkingAccount, err := bankingService.CreateAccount(customer.ID, 1000.0, models.Checking)
	if err != nil {
		fmt.Printf("Error creating checking account: %v\n", err)
		return
	}

	fmt.Printf("Created checking account: ID=%s, Balance=%.2f\n", checkingAccount.ID, checkingAccount.Balance)

	// Create a savings account
	savingsAccount, err := bankingService.CreateAccount(customer.ID, 5000.0, models.Savings)
	if err != nil {
		fmt.Printf("Error creating savings account: %v\n", err)
		return
	}

	fmt.Printf("Created savings account: ID=%s, Balance=%.2f\n", savingsAccount.ID, savingsAccount.Balance)

	// Make a deposit
	err = bankingService.Deposit(checkingAccount.ID, 500.0)
	if err != nil {
		fmt.Printf("Error making deposit: %v\n", err)
		return
	}

	// Get updated account
	checkingAccount, _ = bankingService.GetAccount(checkingAccount.ID)
	fmt.Printf("Deposited $500. New checking balance: $%.2f\n", checkingAccount.Balance)

	// Make a withdrawal
	err = bankingService.Withdraw(savingsAccount.ID, 1000.0)
	if err != nil {
		fmt.Printf("Error making withdrawal: %v\n", err)
		return
	}

	// Get updated account
	savingsAccount, _ = bankingService.GetAccount(savingsAccount.ID)
	fmt.Printf("Withdrew $1000. New savings balance: $%.2f\n", savingsAccount.Balance)

	// Make a transfer
	err = bankingService.Transfer(savingsAccount.ID, checkingAccount.ID, 500.0)
	if err != nil {
		fmt.Printf("Error making transfer: %v\n", err)
		return
	}

	// Get updated accounts
	checkingAccount, _ = bankingService.GetAccount(checkingAccount.ID)
	savingsAccount, _ = bankingService.GetAccount(savingsAccount.ID)

	fmt.Printf("Transferred $500 from savings to checking.\n")
	fmt.Printf("New checking balance: $%.2f\n", checkingAccount.Balance)
	fmt.Printf("New savings balance: $%.2f\n", savingsAccount.Balance)

	// Get transactions for checking account
	transactions, err := bankingService.GetAccountTransactions(checkingAccount.ID)
	if err != nil {
		fmt.Printf("Error getting transactions: %v\n", err)
		return
	}

	fmt.Printf("Transactions for checking account (%d total):\n", len(transactions))
	for i, tx := range transactions {
		fmt.Printf("%d. Type: %s, Amount: $%.2f, Date: %s\n", i+1, tx.TransactionType, tx.Amount, tx.Timestamp.Format("2006-01-02 15:04:05"))
	}

	// Demonstrate adapter pattern by switching notification method
	fmt.Println("\n=== Demonstrating Adapter Pattern ===")
	notificationAdapter.SendNotification(customer.ID, "This is an email notification")
	notificationAdapter.ToggleNotificationMethod() // Switch to SMS
	notificationAdapter.SendNotification(customer.ID, "This is an SMS notification")

	// Demonstrate reporting adapter
	fmt.Println("\n=== Demonstrating External System Integration ===")
	lastTransaction := transactions[len(transactions)-1]

	// Report transaction to external system
	err = reportingAdapter.ReportTransaction(lastTransaction)
	if err != nil {
		fmt.Printf("Error reporting transaction: %v\n", err)
	}

	// Generate a report
	report, err := reportingAdapter.GenerateAccountReport(checkingAccount.ID)
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
	} else {
		fmt.Println(report)
	}

	// Demonstrate generic repository pattern
	fmt.Println("\n=== Demonstrating Generic Repository Pattern ===")

	// Create generic repositories for each entity type
	genericCustomerRepo := adapters.NewGenericRepository[*models.Customer]()
	genericAccountRepo := adapters.NewGenericRepository[*models.Account]()
	genericTransactionRepo := adapters.NewGenericRepository[*models.Transaction]()

	// Add existing entities to the generic repositories
	fmt.Println("Adding entities to generic repositories...")

	// Add customer
	err = genericCustomerRepo.Create(customer)
	if err != nil {
		fmt.Printf("Error adding customer to generic repo: %v\n", err)
	}

	// Add accounts
	err = genericAccountRepo.Create(checkingAccount)
	if err != nil {
		fmt.Printf("Error adding checking account to generic repo: %v\n", err)
	}

	err = genericAccountRepo.Create(savingsAccount)
	if err != nil {
		fmt.Printf("Error adding savings account to generic repo: %v\n", err)
	}

	// Add transactions
	for _, tx := range transactions {
		err = genericTransactionRepo.Create(tx)
		if err != nil {
			fmt.Printf("Error adding transaction to generic repo: %v\n", err)
		}
	}

	// Retrieve and display counts
	fmt.Printf("Generic Customer Repository count: %d\n", genericCustomerRepo.Count())
	fmt.Printf("Generic Account Repository count: %d\n", genericAccountRepo.Count())
	fmt.Printf("Generic Transaction Repository count: %d\n", genericTransactionRepo.Count())

	// Demonstrate filtering with generics
	fmt.Println("\nDemonstrating generic filtering:")

	// Filter accounts by type
	checkingAccounts := genericAccountRepo.Filter(func(a *models.Account) bool {
		return a.AccountType == models.Checking
	})

	fmt.Printf("Number of checking accounts: %d\n", len(checkingAccounts))

	// Filter transactions by type
	depositTransactions := genericTransactionRepo.Filter(func(t *models.Transaction) bool {
		return t.TransactionType == models.Deposit
	})

	fmt.Printf("Number of deposit transactions: %d\n", len(depositTransactions))

	fmt.Println("\n=== Simulation Complete ===")
}
