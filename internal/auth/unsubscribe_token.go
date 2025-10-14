package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UnsubscribeTokenGenerator generates and validates unsubscribe tokens
type UnsubscribeTokenGenerator struct {
	secretKey []byte
}

// NewUnsubscribeTokenGenerator creates a new token generator with the given secret
func NewUnsubscribeTokenGenerator(secretKey []byte) *UnsubscribeTokenGenerator {
	return &UnsubscribeTokenGenerator{
		secretKey: secretKey,
	}
}

// GenerateToken creates an unsubscribe token for the given email and expiry
func (g *UnsubscribeTokenGenerator) GenerateToken(email string, expiresAt time.Time) string {
	// create payload: email + expiry timestamp
	payload := fmt.Sprintf("%s:%d", email, expiresAt.Unix())

	// create HMAC signature
	h := hmac.New(sha256.New, g.secretKey)
	h.Write([]byte(payload))
	signature := h.Sum(nil)

	// combine signature + payload and encode
	tokenData := make([]byte, 0, len(signature)+len(payload))
	tokenData = append(tokenData, signature...)
	tokenData = append(tokenData, []byte(payload)...)
	return base64.URLEncoding.EncodeToString(tokenData)
}

// ValidateToken validates an unsubscribe token and returns the email if valid
func (g *UnsubscribeTokenGenerator) ValidateToken(token string) (string, error) {
	// decode token
	tokenData, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", errors.New("invalid token format")
	}

	if len(tokenData) < 32 {
		return "", errors.New("token too short")
	}

	// extract signature and payload
	signature := tokenData[:32]
	payload := string(tokenData[32:])

	// verify HMAC signature
	h := hmac.New(sha256.New, g.secretKey)
	h.Write([]byte(payload))
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return "", errors.New("invalid token signature")
	}

	// parse payload
	parts := strings.Split(payload, ":")
	if len(parts) != 2 {
		return "", errors.New("invalid token payload")
	}

	email := parts[0]
	expiryStr := parts[1]

	// parse expiry time
	expiry, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return "", errors.New("invalid expiry time")
	}

	// check if current time is past expiry
	if time.Now().Unix() > expiry {
		return "", errors.New("token expired")
	}

	return email, nil
}
