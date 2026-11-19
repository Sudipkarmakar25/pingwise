package auth_service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const (
	accessTokenExpiry  = 30 * 24 * time.Hour
	refreshTokenExpiry = 60 * 24 * time.Hour
)

// generateToken returns a random opaque token. Only its hash is stored.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
