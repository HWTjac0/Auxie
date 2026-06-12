package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateRandomString tworzy bezpieczny ciąg losowych znaków o długości n
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}

// GeneratePKCE tworzy parę (verifier, challenge) niezbędną w nowoczesnym flow OAuth2
func GeneratePKCE() (string, string, error) {
	verifier, err := GenerateRandomString(64)
	if err != nil {
		return "", "", err
	}

	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}
