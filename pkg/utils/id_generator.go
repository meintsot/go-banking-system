package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// UUIDGenerator implements IDGenerator for creating unique IDs
type UUIDGenerator struct {
	r *rand.Rand
}

// NewUUIDGenerator creates a new UUID generator with randomized seed
func NewUUIDGenerator() *UUIDGenerator {
	seed := rand.NewSource(time.Now().UnixNano())
	return &UUIDGenerator{
		r: rand.New(seed),
	}
}

// GenerateID creates a new unique ID
func (g *UUIDGenerator) GenerateID() string {
	// Simple UUID generation for demonstration purposes
	timestamp := time.Now().UnixNano()
	randomNum := g.r.Int63()
	return fmt.Sprintf("%x-%x", timestamp, randomNum)
}
