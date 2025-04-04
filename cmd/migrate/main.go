package main

import (
	"bankingsystem/pkg/database"
	"bankingsystem/pkg/utils"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func main() {
	// Define command line flags
	sourceDB := flag.String("source", "", "Path to the source database (old format)")
	targetDB := flag.String("target", "banking.db", "Path to the target database (GORM format)")
	flag.Parse()

	// If source is not provided, try to find the latest backup
	if *sourceDB == "" {
		backups, err := filepath.Glob("banking_backup_*.db")
		if err != nil {
			log.Fatalf("Error finding backup databases: %v", err)
		}

		if len(backups) == 0 {
			log.Fatalf("No source database specified and no backup found. Please use -source flag.")
		}

		// Find the latest backup using regex to extract the timestamp
		latestBackup := backups[0]
		var latestTime time.Time

		// Regex to extract date from backup filenames (banking_backup_YYYYMMDD_HHMMSS.db)
		re := regexp.MustCompile(`banking_backup_(\d{8}_\d{6})\.db`)

		for _, backup := range backups {
			matches := re.FindStringSubmatch(backup)
			if len(matches) != 2 {
				log.Printf("Warning: Could not parse time from backup filename %s", backup)
				continue
			}

			dateStr := matches[1]
			backupTime, err := time.ParseInLocation("20060102_150405", dateStr, time.Local)
			if err != nil {
				log.Printf("Warning: Could not parse time from backup filename %s: %v", backup, err)
				continue
			}

			if backupTime.After(latestTime) {
				latestTime = backupTime
				latestBackup = backup
			}
		}

		*sourceDB = latestBackup
		log.Printf("Using latest backup database: %s", *sourceDB)
	}

	// Create logger
	logger := log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags)

	logger.Printf("Starting database migration from %s to %s", *sourceDB, *targetDB)

	// Initialize GORM connection for the target database
	dbConn, err := database.NewGormDBConnection(*targetDB)
	if err != nil {
		log.Fatalf("Failed to initialize target database: %v", err)
	}
	defer dbConn.Close()

	// Initialize database schema
	logger.Println("Initializing target database schema...")
	if err := dbConn.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Create and run the migrator
	migrator, err := utils.NewDBMigrator(*sourceDB, dbConn.GetDB(), logger)
	if err != nil {
		log.Fatalf("Failed to create database migrator: %v", err)
	}
	defer migrator.Close()

	// Perform the migration
	startTime := time.Now()
	logger.Println("Starting data migration...")

	if err := migrator.MigrateAll(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	elapsedTime := time.Since(startTime)
	logger.Printf("Migration completed successfully in %v", elapsedTime)

	fmt.Println("\nPlease check the logs above to verify the migration was successful.")
	fmt.Println("If everything looks good, you can now run the main application with the migrated data.")
}
