package utils

import (
	"fmt"
	"io"
	"os"
)

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open source file: %v", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("could not create destination file: %v", err)
	}
	defer destFile.Close()

	// Copy contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying file contents: %v", err)
	}

	// Sync to ensure all data is written to disk
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("error syncing file: %v", err)
	}

	return nil
}
