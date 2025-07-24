package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

// KeyGen generates private/public keypair and stores it in files
func KeyGen() error {
	// generate key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// construct PEM block for private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// create file for private key in PEM format
	privateKeyFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private key file: %w", err)
	}
	defer func() {
		if cErr := privateKeyFile.Close(); cErr != nil {
			log.Printf("can't close private key file: %v", cErr)
		}
	}()

	// write private key to file
	if err = pem.Encode(privateKeyFile, &privateKeyBlock); err != nil {
		return fmt.Errorf("writing private key file: %w", err)
	}

	// construct PEM block for public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshal public key to PKIX: %w", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	// create file for public key in PEM format
	publicKeyFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public key file: %w", err)
	}
	defer func() {
		if cErr := publicKeyFile.Close(); cErr != nil {
			log.Printf("can't close public key file: %v", cErr)
		}
	}()

	// write private key to file
	if err = pem.Encode(publicKeyFile, publicKeyBlock); err != nil {
		return fmt.Errorf("writing public key file: %w", err)
	}
	return nil
}
