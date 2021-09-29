package auth

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	algorithm  string
	privateKey *rsa.PrivateKey
	parser     *jwt.Parser
	keyFunc    jwt.Keyfunc
}

// New constructs Auth instance
func New(algorithm string, privateKey *rsa.PrivateKey) (*Auth, error) {
	if jwt.GetSigningMethod(algorithm) == nil {
		return nil, fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	var keyFunc jwt.Keyfunc = func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}

	parser := &jwt.Parser{
		ValidMethods: []string{algorithm},
	}

	a := Auth{
		algorithm:  algorithm,
		privateKey: privateKey,
		parser:     parser,
		keyFunc:    keyFunc,
	}

	return &a, nil
}

// GenerateToken generates JWT token with claims
func (a *Auth) GenerateToken(claims jwt.Claims) (string, error) {
	method := jwt.GetSigningMethod(a.algorithm)

	token := jwt.NewWithClaims(method, claims)

	tokenStr, err := token.SignedString(a.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing crypto: %w", err)
	}

	return tokenStr, nil
}

// ValidateToken validates token and returns claims from it
func (a *Auth) ValidateToken(tokenStr string) (jwt.Claims, error) {
	var claims jwt.MapClaims
	token, err := a.parser.ParseWithClaims(tokenStr, claims, a.keyFunc)
	if err != nil {
		return jwt.MapClaims{}, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return jwt.MapClaims{}, fmt.Errorf("invalid token")
	}

	return claims, nil
}
