package notification_test

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/notification"
	"github.com/stretchr/testify/require"
)

func TestCipherAuthenticatesPayloadAndIdentity(t *testing.T) {
	c, err := notification.NewCipher(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	require.NoError(t, err)
	want := notification.Email{From: "sender@example.com", To: []string{"user@example.com"}, Subject: "Verify", HTML: "code 123456", Text: "123456"}
	sealed, err := c.Encrypt("id", notification.VerificationKind, "verification", want)
	require.NoError(t, err)
	require.NotContains(t, string(sealed), "123456")
	got, err := c.Decrypt("id", notification.VerificationKind, "verification", sealed)
	require.NoError(t, err)
	require.Equal(t, want, got)
	_, err = c.Decrypt("id", "other_kind", "verification", sealed)
	require.Error(t, err)
	other, err := c.Encrypt("id", notification.VerificationKind, "verification", want)
	require.NoError(t, err)
	require.NotEqual(t, sealed, other)
	_, err = c.Decrypt("other", notification.VerificationKind, "verification", sealed)
	require.Error(t, err)
	_, err = c.Decrypt("id", notification.VerificationKind, "other", sealed)
	require.Error(t, err)
	sealed[len(sealed)-1] ^= 1
	_, err = c.Decrypt("id", notification.VerificationKind, "verification", sealed)
	require.Error(t, err)
}

func TestCipherRejectsInvalidKeysAndEnvelopes(t *testing.T) {
	for _, key := range []string{"", "bad-base64", base64.StdEncoding.EncodeToString(make([]byte, 16))} {
		_, err := notification.NewCipher(key)
		require.Error(t, err)
	}
	c, err := notification.NewCipher(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	require.NoError(t, err)
	for _, value := range [][]byte{nil, {1}, {99, 2, 3}} {
		_, err = c.Decrypt("id", notification.VerificationKind, "verification", value)
		require.Error(t, err)
	}
	other, err := notification.NewCipher(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)))
	require.NoError(t, err)
	sealed, err := other.Encrypt("id", notification.VerificationKind, "verification", notification.Email{})
	require.NoError(t, err)
	_, err = c.Decrypt("id", notification.VerificationKind, "verification", sealed)
	require.Error(t, err)
}
