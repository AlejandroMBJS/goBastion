package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Claims represents JWT claims
type Claims struct {
	Sub  string `json:"sub"`  // Subject (user ID)
	Role string `json:"role"` // User role
	Exp  int64  `json:"exp"`  // Expiration time
	Iat  int64  `json:"iat"`  // Issued at
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// GenerateToken creates a JWT token using HS256
func GenerateToken(secret, sub, role string, ttlMinutes int) (string, error) {
	now := time.Now().Unix()
	claims := Claims{
		Sub:  sub,
		Role: role,
		Iat:  now,
		Exp:  now + int64(ttlMinutes*60),
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// Base64 URL encode
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature := createSignature(message, secret)

	// Combine all parts
	token := message + "." + signature

	return token, nil
}

// ParseAndValidateToken parses and validates a JWT token
func ParseAndValidateToken(secret, token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	headerEncoded := parts[0]
	claimsEncoded := parts[1]
	signatureEncoded := parts[2]

	// Verify signature
	message := headerEncoded + "." + claimsEncoded
	expectedSignature := createSignature(message, secret)
	if signatureEncoded != expectedSignature {
		return Claims{}, ErrInvalidToken
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}

	// Check expiration
	now := time.Now().Unix()
	if claims.Exp < now {
		return Claims{}, ErrExpiredToken
	}

	return claims, nil
}

// createSignature creates an HMAC-SHA256 signature
func createSignature(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}

// ExtractBearerToken extracts the token from Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// FormatTokenResponse creates a standard token response
func FormatTokenResponse(accessToken, refreshToken string) map[string]any {
	response := map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
	}
	if refreshToken != "" {
		response["refresh_token"] = refreshToken
	}
	return response
}

// ClaimsToUserID converts claims subject to user ID string
func ClaimsToUserID(claims Claims) string {
	return claims.Sub
}

// UserIDToClaims creates claims from user ID and role
func UserIDToClaims(userID, role string, ttlMinutes int) Claims {
	now := time.Now().Unix()
	return Claims{
		Sub:  userID,
		Role: role,
		Iat:  now,
		Exp:  now + int64(ttlMinutes*60),
	}
}

// ValidateTokenString is a convenience function that validates a token string
func ValidateTokenString(secret, authHeader string) (Claims, error) {
	token, err := ExtractBearerToken(authHeader)
	if err != nil {
		return Claims{}, err
	}
	return ParseAndValidateToken(secret, token)
}

// CreateTokenPair creates both access and refresh tokens
func CreateTokenPair(secret, sub, role string, accessMinutes, refreshMinutes int) (string, string, error) {
	accessToken, err := GenerateToken(secret, sub, role, accessMinutes)
	if err != nil {
		return "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := GenerateToken(secret, sub, role, refreshMinutes)
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
