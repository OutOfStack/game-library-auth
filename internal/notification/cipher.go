// Package notification contains the immutable email payload and its encrypted envelope
package notification

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// VerificationKind authenticates the supported notification payload type
const VerificationKind = "email_verification"

// ErrRejected marks a delivery the provider refused permanently, so retrying cannot help
var ErrRejected = errors.New("email rejected by provider")

// ErrMisconfigured marks a delivery refused because of sender credentials or configuration, so it can succeed once fixed
var ErrMisconfigured = errors.New("email provider rejected sender credentials or configuration")

// Email is rendered once, before enqueueing, so retries have identical provider payloads
type Email struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

// Cipher encrypts outbox payloads using a shared AES-256-GCM key
type Cipher struct{ aead cipher.AEAD }

// NewCipher validates and decodes the shared base64 key
func NewCipher(key string) (*Cipher, error) {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(decoded) != 32 {
		return nil, errors.New("NOTIFICATION_OUTBOX_KEY must be base64-encoded 32 bytes")
	}
	block, err := aes.NewCipher(decoded)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt binds email content to its notification kind and reference identities
func (c *Cipher) Encrypt(id, kind, referenceID string, email Email) ([]byte, error) {
	plain, err := json.Marshal(email)
	if err != nil {
		return nil, err
	}
	// envelope: version byte, random nonce, authenticated ciphertext
	envelope := make([]byte, 1+c.aead.NonceSize())
	envelope[0] = 1
	if _, err = rand.Read(envelope[1:]); err != nil {
		return nil, fmt.Errorf("generate notification nonce: %w", err)
	}
	return c.aead.Seal(envelope, envelope[1:], plain, identity(id, kind, referenceID)), nil
}

// Decrypt rejects tampered envelopes and payloads copied from another notification
func (c *Cipher) Decrypt(id, kind, referenceID string, envelope []byte) (Email, error) {
	start := 1 + c.aead.NonceSize()
	if len(envelope) < start+c.aead.Overhead() || envelope[0] != 1 {
		return Email{}, errors.New("invalid notification envelope")
	}
	plain, err := c.aead.Open(nil, envelope[1:start], envelope[start:], identity(id, kind, referenceID))
	if err != nil {
		return Email{}, errors.New("cannot authenticate notification payload; check the shared outbox key")
	}
	var email Email
	if err = json.Unmarshal(plain, &email); err != nil {
		return Email{}, errors.New("invalid notification payload")
	}
	return email, nil
}

func identity(id, kind, referenceID string) []byte {
	return []byte(kind + "\x00" + id + "\x00" + referenceID)
}
