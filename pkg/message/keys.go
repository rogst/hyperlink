package message

import (
	"math/rand"
	"time"
)

const (
	// KeyLetters is the characters used to generate message keys
	keyLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewKeyFromRandomLetters returns a key made from random letters of the specified length
func NewKeyFromRandomLetters(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = keyLetters[rand.Intn(len(keyLetters))]
	}

	return string(b)
}
