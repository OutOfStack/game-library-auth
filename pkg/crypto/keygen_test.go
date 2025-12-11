package crypto_test

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/OutOfStack/game-library-auth/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyGen(t *testing.T) {
	t.Run("successfully generates private and public key files", func(t *testing.T) {
		tempDir := t.TempDir()

		t.Chdir(tempDir)

		err := crypto.KeyGen()
		require.NoError(t, err)

		privateKeyData, err := os.ReadFile("private.pem")
		require.NoError(t, err)
		assert.NotEmpty(t, privateKeyData)

		publicKeyData, err := os.ReadFile("public.pem")
		require.NoError(t, err)
		assert.NotEmpty(t, publicKeyData)

		privateBlock, _ := pem.Decode(privateKeyData)
		require.NotNil(t, privateBlock)
		assert.Equal(t, "RSA PRIVATE KEY", privateBlock.Type)

		privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
		require.NoError(t, err)
		assert.NotNil(t, privateKey)
		assert.Equal(t, 2048, privateKey.N.BitLen())

		publicBlock, _ := pem.Decode(publicKeyData)
		require.NotNil(t, publicBlock)
		assert.Equal(t, "RSA PUBLIC KEY", publicBlock.Type)

		publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
		require.NoError(t, err)
		assert.NotNil(t, publicKey)
	})

	t.Run("overwrites existing key files", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Chdir(tempDir)

		err := os.WriteFile("private.pem", []byte("old private key"), 0600)
		require.NoError(t, err)
		err = os.WriteFile("public.pem", []byte("old public key"), 0600)
		require.NoError(t, err)

		err = crypto.KeyGen()
		require.NoError(t, err)

		privateKeyData, err := os.ReadFile("private.pem")
		require.NoError(t, err)
		assert.NotEqual(t, "old private key", string(privateKeyData))

		publicKeyData, err := os.ReadFile("public.pem")
		require.NoError(t, err)
		assert.NotEqual(t, "old public key", string(publicKeyData))

		privateBlock, _ := pem.Decode(privateKeyData)
		require.NotNil(t, privateBlock)
		assert.Equal(t, "RSA PRIVATE KEY", privateBlock.Type)
	})
}

func TestGenerateSecret(t *testing.T) {
	t.Run("generates secret of correct length", func(t *testing.T) {
		secret, err := crypto.GenerateSecret(32)
		require.NoError(t, err)
		assert.NotEmpty(t, secret)
	})

	t.Run("generates unique secrets each time", func(t *testing.T) {
		secret1, err := crypto.GenerateSecret(32)
		require.NoError(t, err)

		secret2, err := crypto.GenerateSecret(32)
		require.NoError(t, err)

		assert.NotEqual(t, secret1, secret2)
	})

	t.Run("generates secrets of different lengths", func(t *testing.T) {
		secret16, err := crypto.GenerateSecret(16)
		require.NoError(t, err)

		secret32, err := crypto.GenerateSecret(32)
		require.NoError(t, err)

		secret64, err := crypto.GenerateSecret(64)
		require.NoError(t, err)

		assert.NotEmpty(t, secret16)
		assert.NotEmpty(t, secret32)
		assert.NotEmpty(t, secret64)
		assert.NotEqual(t, len(secret16), len(secret32))
		assert.NotEqual(t, len(secret32), len(secret64))
	})

	t.Run("handles zero length", func(t *testing.T) {
		secret, err := crypto.GenerateSecret(0)
		require.NoError(t, err)
		assert.Empty(t, secret)
	})

	t.Run("generates valid base64 encoded string", func(t *testing.T) {
		secret, err := crypto.GenerateSecret(32)
		require.NoError(t, err)
		assert.Regexp(t, "^[A-Za-z0-9+/]+=*$", secret)
	})
}
