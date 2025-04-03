package adapters

import (
	"bankingsystem/pkg/models"
	"errors"
	"fmt"
	"time"
)

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendNotification(recipientID, message string) error
}

// ExternalReportingSystem defines the interface for reporting to an external system
type ExternalReportingSystem interface {
	ReportTransaction(transaction *models.Transaction) error
	GenerateAccountReport(accountID string) (string, error)
}

// EmailNotificationService implements NotificationService for email notifications
type EmailNotificationService struct {
	// In a real system, this might contain an email client configuration
}

// NewEmailNotificationService creates a new email notification service
func NewEmailNotificationService() *EmailNotificationService {
	return &EmailNotificationService{}
}

// SendNotification sends an email notification
func (s *EmailNotificationService) SendNotification(recipientID, message string) error {
	// In a real system, this would send an actual email
	fmt.Printf("[Email] Sending notification to %s: %s\n", recipientID, message)
	return nil
}

// SMSNotificationService implements NotificationService for SMS notifications
type SMSNotificationService struct {
	// In a real system, this might contain an SMS gateway configuration
}

// NewSMSNotificationService creates a new SMS notification service
func NewSMSNotificationService() *SMSNotificationService {
	return &SMSNotificationService{}
}

// SendNotification sends an SMS notification
func (s *SMSNotificationService) SendNotification(recipientID, message string) error {
	// In a real system, this would send an actual SMS
	fmt.Printf("[SMS] Sending notification to %s: %s\n", recipientID, message)
	return nil
}

// NotificationAdapter is an adapter that can switch between different notification services
type NotificationAdapter struct {
	emailService *EmailNotificationService
	smsService   *SMSNotificationService
	useEmail     bool
}

// NewNotificationAdapter creates a new notification adapter
func NewNotificationAdapter(useEmail bool) *NotificationAdapter {
	return &NotificationAdapter{
		emailService: NewEmailNotificationService(),
		smsService:   NewSMSNotificationService(),
		useEmail:     useEmail,
	}
}

// SendNotification sends a notification using the selected service
func (a *NotificationAdapter) SendNotification(recipientID, message string) error {
	if a.useEmail {
		return a.emailService.SendNotification(recipientID, message)
	}
	return a.smsService.SendNotification(recipientID, message)
}

// ToggleNotificationMethod switches between email and SMS
func (a *NotificationAdapter) ToggleNotificationMethod() {
	a.useEmail = !a.useEmail
}

// MockExternalReportingSystem implements ExternalReportingSystem for reporting
type MockExternalReportingSystem struct {
	// In a real system, this might contain API credentials or configuration
	reportedTransactions map[string]*models.Transaction
}

// NewMockExternalReportingSystem creates a new mock external reporting system
func NewMockExternalReportingSystem() *MockExternalReportingSystem {
	return &MockExternalReportingSystem{
		reportedTransactions: make(map[string]*models.Transaction),
	}
}

// ReportTransaction reports a transaction to the external system
func (s *MockExternalReportingSystem) ReportTransaction(transaction *models.Transaction) error {
	// In a real system, this would make an API call to an external system
	fmt.Printf("[External System] Reporting transaction %s: Amount = %.2f, Type = %s\n",
		transaction.ID, transaction.Amount, transaction.TransactionType)

	s.reportedTransactions[transaction.ID] = transaction
	return nil
}

// GenerateAccountReport generates a report for an account
func (s *MockExternalReportingSystem) GenerateAccountReport(accountID string) (string, error) {
	// In a real system, this would generate a real report from external data
	count := 0
	for _, t := range s.reportedTransactions {
		if t.AccountID == accountID || t.DestinationAccountID == accountID {
			count++
		}
	}

	if count == 0 {
		return "", errors.New("no transactions found for this account")
	}

	return fmt.Sprintf("[External System] Generated report for account %s at %s with %d transactions",
		accountID, time.Now().Format(time.RFC3339), count), nil
}

// ReportingAdapter is an adapter for working with different reporting systems
type ReportingAdapter struct {
	reportingSystem ExternalReportingSystem
}

// NewReportingAdapter creates a new reporting adapter
func NewReportingAdapter(reportingSystem ExternalReportingSystem) *ReportingAdapter {
	return &ReportingAdapter{
		reportingSystem: reportingSystem,
	}
}

// ReportTransaction reports a transaction to the current reporting system
func (a *ReportingAdapter) ReportTransaction(transaction *models.Transaction) error {
	return a.reportingSystem.ReportTransaction(transaction)
}

// GenerateAccountReport generates a report for an account
func (a *ReportingAdapter) GenerateAccountReport(accountID string) (string, error) {
	return a.reportingSystem.GenerateAccountReport(accountID)
}
