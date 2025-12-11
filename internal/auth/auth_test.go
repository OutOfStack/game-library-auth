package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateValidate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	a, err := auth.New("RS256", privateKey, "", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test_runner",
			Subject:   "12345qwerty",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(720 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserRole: "super_admin",
	}

	tokenStr, err := a.GenerateToken(claims)
	require.NoError(t, err)

	parsedClaims, err := a.GetClaimsFromToken(tokenStr)
	require.NoError(t, err)

	assert.Equal(t, claims.Issuer, parsedClaims.Issuer)
	assert.Equal(t, claims.Subject, parsedClaims.Subject)
	assert.Equal(t, claims.UserRole, parsedClaims.UserRole)
}

func TestGenerateRefreshToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	a, err := auth.New("RS256", privateKey, "", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	token1, _, err := a.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, _, err := a.GenerateRefreshToken()
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2)
	assert.GreaterOrEqual(t, len(token1), 40)
}

func TestNew(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tests := []struct {
		name            string
		algorithm       string
		claimsIssuer    string
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
		wantErr         bool
	}{
		{
			name:            "valid RS256 algorithm",
			algorithm:       "RS256",
			claimsIssuer:    "test-issuer",
			accessTokenTTL:  15 * time.Minute,
			refreshTokenTTL: 7 * 24 * time.Hour,
			wantErr:         false,
		},
		{
			name:            "valid RS384 algorithm",
			algorithm:       "RS384",
			claimsIssuer:    "test-issuer",
			accessTokenTTL:  15 * time.Minute,
			refreshTokenTTL: 7 * 24 * time.Hour,
			wantErr:         false,
		},
		{
			name:            "valid RS512 algorithm",
			algorithm:       "RS512",
			claimsIssuer:    "test-issuer",
			accessTokenTTL:  15 * time.Minute,
			refreshTokenTTL: 7 * 24 * time.Hour,
			wantErr:         false,
		},
		{
			name:            "invalid algorithm",
			algorithm:       "INVALID",
			claimsIssuer:    "test-issuer",
			accessTokenTTL:  15 * time.Minute,
			refreshTokenTTL: 7 * 24 * time.Hour,
			wantErr:         true,
		},
		{
			name:            "empty algorithm",
			algorithm:       "",
			claimsIssuer:    "test-issuer",
			accessTokenTTL:  15 * time.Minute,
			refreshTokenTTL: 7 * 24 * time.Hour,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := auth.New(tt.algorithm, privateKey, tt.claimsIssuer, tt.accessTokenTTL, tt.refreshTokenTTL)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, a)
			}
		})
	}
}

func TestGetClaimsFromToken_Errors(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	t.Run("expired token", func(t *testing.T) {
		claims := auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "test-issuer",
				Subject:   "user123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
			UserID: "user123",
		}

		tokenStr, err := a.GenerateToken(claims)
		require.NoError(t, err)

		_, err = a.GetClaimsFromToken(tokenStr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := a.GetClaimsFromToken("malformed.token.string")
		require.Error(t, err)
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := a.GetClaimsFromToken("")
		require.Error(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		otherAuth, err := auth.New("RS256", otherKey, "other-issuer", 15*time.Minute, 7*24*time.Hour)
		require.NoError(t, err)

		claims := auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "other-issuer",
				Subject:   "user123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID: "user123",
		}

		tokenStr, err := otherAuth.GenerateToken(claims)
		require.NoError(t, err)

		_, err = a.GetClaimsFromToken(tokenStr)
		require.Error(t, err)
	})

	t.Run("token not yet valid", func(t *testing.T) {
		claims := auth.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "test-issuer",
				Subject:   "user123",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserID: "user123",
		}

		tokenStr, err := a.GenerateToken(claims)
		require.NoError(t, err)

		_, err = a.GetClaimsFromToken(tokenStr)
		require.Error(t, err)
	})
}

func TestGenerateRefreshToken_ExpiresAt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	refreshTokenTTL := 7 * 24 * time.Hour
	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, refreshTokenTTL)
	require.NoError(t, err)

	beforeGeneration := time.Now()
	token, expiresAt, err := a.GenerateRefreshToken()
	afterGeneration := time.Now()

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	expectedMinExpiry := beforeGeneration.Add(refreshTokenTTL)
	expectedMaxExpiry := afterGeneration.Add(refreshTokenTTL)

	assert.True(t, expiresAt.After(expectedMinExpiry) || expiresAt.Equal(expectedMinExpiry))
	assert.True(t, expiresAt.Before(expectedMaxExpiry) || expiresAt.Equal(expectedMaxExpiry))
}

func TestGenerateToken_AllClaims(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	a, err := auth.New("RS256", privateKey, "test-issuer", 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   "user123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   "user123",
		UserRole: "admin",
		Username: "testuser",
		Name:     "Test User",
	}

	tokenStr, err := a.GenerateToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	parsedClaims, err := a.GetClaimsFromToken(tokenStr)
	require.NoError(t, err)

	assert.Equal(t, claims.UserID, parsedClaims.UserID)
	assert.Equal(t, claims.UserRole, parsedClaims.UserRole)
	assert.Equal(t, claims.Username, parsedClaims.Username)
	assert.Equal(t, claims.Name, parsedClaims.Name)
}
