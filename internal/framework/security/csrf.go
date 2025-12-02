package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// GenerateCSRFToken generates a random CSRF token
func GenerateCSRFToken(secret string) (string, error) {
	// Generate random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(b)
	signature := h.Sum(nil)

	// Combine random bytes and signature
	token := append(b, signature...)

	// Base64 URL encode
	return base64.RawURLEncoding.EncodeToString(token), nil
}

// ValidateCSRFToken validates a CSRF token using double-submit cookie pattern
func ValidateCSRFToken(secret, cookieToken, headerToken string) bool {
	// Both tokens must be present
	if cookieToken == "" || headerToken == "" {
		return false
	}

	// Tokens must match (double-submit)
	if cookieToken != headerToken {
		return false
	}

	// Decode the token
	decoded, err := base64.RawURLEncoding.DecodeString(cookieToken)
	if err != nil {
		return false
	}

	// Must be at least 32 bytes (random) + 32 bytes (HMAC-SHA256)
	if len(decoded) < 64 {
		return false
	}

	// Split into random bytes and signature
	randomBytes := decoded[:32]
	providedSignature := decoded[32:]

	// Recreate signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(randomBytes)
	expectedSignature := h.Sum(nil)

	// Compare signatures
	return hmac.Equal(expectedSignature, providedSignature)
}

// GenerateSimpleToken generates a simple random token (without HMAC)
// Useful for scenarios where the secret validation is not required
func GenerateSimpleToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ValidateSimpleToken validates a simple token (checks if not empty and properly encoded)
func ValidateSimpleToken(token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	// Try to decode to ensure it's valid base64
	_, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return errors.New("invalid token encoding")
	}

	return nil
}
