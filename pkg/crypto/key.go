package crypto

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v4"
)

// ReadPrivateKey reads private key from file path
func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	privatePEM, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	return privateKey, nil
}
