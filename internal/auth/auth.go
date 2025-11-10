package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Auth represents dependencies for auth methods
type Auth struct {
	algorithm       string
	privateKey      *rsa.PrivateKey
	parser          *jwt.Parser
	keyFunc         jwt.Keyfunc
	claimsIssuer    string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// New constructs Auth instance
func New(algorithm string, privateKey *rsa.PrivateKey, claimsIssuer string, accessTokenTTL, refreshTokenTTL time.Duration) (*Auth, error) {
	if jwt.GetSigningMethod(algorithm) == nil {
		return nil, fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	var keyFunc jwt.Keyfunc = func(_ *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}

	a := Auth{
		algorithm:       algorithm,
		privateKey:      privateKey,
		parser:          jwt.NewParser(jwt.WithValidMethods([]string{algorithm})),
		keyFunc:         keyFunc,
		claimsIssuer:    claimsIssuer,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}

	return &a, nil
}

// GenerateToken generates JWT token with claims
func (a *Auth) GenerateToken(claims jwt.Claims) (string, error) {
	method := jwt.GetSigningMethod(a.algorithm)

	token := jwt.NewWithClaims(method, claims)

	tokenStr, err := token.SignedString(a.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return tokenStr, nil
}

// ValidateToken validates token and returns claims from it
func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)
	if err != nil {
		return Claims{}, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	active := claims.VerifyExpiresAt(time.Now(), true)
	if !active {
		return Claims{}, errors.New("token expired")
	}

	return claims, nil
}

// GenerateRefreshToken returns generated random refresh token and its expiration date
func (a *Auth) GenerateRefreshToken() (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, fmt.Errorf("generating random bytes: %w", err)
	}
	expiresAt := time.Now().Add(a.refreshTokenTTL)
	return base64.URLEncoding.EncodeToString(b), expiresAt, nil
}
