package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

// UUIDGenerator generates UUIDs for entity IDs
type UUIDGenerator struct{}

// NewUUIDGenerator creates a new UUID generator
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// GenerateID generates a new UUID
func (g *UUIDGenerator) GenerateID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		// Fallback to a timestamp-based ID if random generation fails
		return fmt.Sprintf("id-%d", getTimestampMicros())
	}

	// Set version (4) and variant (RFC4122)
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// getTimestampMicros returns the current timestamp in microseconds
func getTimestampMicros() int64 {
	// Use the system time in nanoseconds and convert to microseconds
	return time.Now().UnixNano() / 1000
}
