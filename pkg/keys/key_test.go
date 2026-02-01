package keys_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/OutOfStack/game-library-auth/pkg/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPrivateKey(t *testing.T) {
	t.Run("successfully reads valid private key", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		tempDir := t.TempDir()
		keyPath := filepath.Join(tempDir, "test_private.pem")

		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		}

		privateKeyFile, err := os.Create(keyPath)
		require.NoError(t, err)

		err = pem.Encode(privateKeyFile, privateKeyBlock)
		require.NoError(t, err)
		require.NoError(t, privateKeyFile.Close())

		readKey, err := keys.ReadPrivateKey(keyPath)
		require.NoError(t, err)
		assert.NotNil(t, readKey)
		assert.Equal(t, privateKey.D, readKey.D)
		assert.Equal(t, privateKey.N, readKey.N)
	})

	t.Run("returns error when file does not exist", func(t *testing.T) {
		key, err := keys.ReadPrivateKey("/nonexistent/path/private.pem")
		require.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "reading private key file")
	})

	t.Run("returns error when file contains invalid PEM", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := filepath.Join(tempDir, "invalid.pem")

		err := os.WriteFile(keyPath, []byte("invalid pem content"), 0600)
		require.NoError(t, err)

		key, err := keys.ReadPrivateKey(keyPath)
		require.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "parsing private key")
	})

	t.Run("returns error when file contains public key instead of private", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		tempDir := t.TempDir()
		keyPath := filepath.Join(tempDir, "public.pem")

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)

		publicKeyBlock := &pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		}

		publicKeyFile, err := os.Create(keyPath)
		require.NoError(t, err)

		err = pem.Encode(publicKeyFile, publicKeyBlock)
		require.NoError(t, err)
		require.NoError(t, publicKeyFile.Close())

		key, err := keys.ReadPrivateKey(keyPath)
		require.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "parsing private key")
	})

	t.Run("returns error when file is empty", func(t *testing.T) {
		tempDir := t.TempDir()
		keyPath := filepath.Join(tempDir, "empty.pem")

		err := os.WriteFile(keyPath, []byte{}, 0600)
		require.NoError(t, err)

		key, err := keys.ReadPrivateKey(keyPath)
		require.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "parsing private key")
	})
}
