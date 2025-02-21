package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Auth represents dependencies for auth methods
type Auth struct {
	algorithm    string
	privateKey   *rsa.PrivateKey
	parser       *jwt.Parser
	keyFunc      jwt.Keyfunc
	claimsIssuer string
}

// New constructs Auth instance
func New(algorithm string, privateKey *rsa.PrivateKey, claimsIssuer string) (*Auth, error) {
	if jwt.GetSigningMethod(algorithm) == nil {
		return nil, fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	var keyFunc jwt.Keyfunc = func(_ *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}

	a := Auth{
		algorithm:    algorithm,
		privateKey:   privateKey,
		parser:       jwt.NewParser(jwt.WithValidMethods([]string{algorithm})),
		keyFunc:      keyFunc,
		claimsIssuer: claimsIssuer,
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
