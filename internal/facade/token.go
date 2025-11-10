package facade

import (
	"context"
	"errors"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/OutOfStack/game-library-auth/internal/model"
	"go.uber.org/zap"
)

var (
	// ErrRefreshTokenNotFound is returned when refresh token is not found
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	// ErrRefreshTokenExpired is returned when refresh token has expired
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

// CreateTokens creates access token and refresh token for a user
func (p *Provider) CreateTokens(ctx context.Context, user model.User) (TokenPair, error) {
	// create access token
	claims := p.auth.CreateUserClaims(user)
	accessToken, err := p.auth.GenerateToken(claims)
	if err != nil {
		p.log.Error("generate access token", zap.String("userID", user.ID), zap.Error(err))
		return TokenPair{}, err
	}

	// create refresh token
	refreshToken, err := p.CreateRefreshToken(ctx, user.ID)
	if err != nil {
		p.log.Error("create refresh token", zap.String("userID", user.ID), zap.Error(err))
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// CreateRefreshToken creates a refresh token for the user
func (p *Provider) CreateRefreshToken(ctx context.Context, userID string) (RefreshToken, error) {
	refreshTokenStr, expiresAt, err := p.auth.GenerateRefreshToken()
	if err != nil {
		p.log.Error("generate refresh token", zap.String("userID", userID), zap.Error(err))
		return RefreshToken{}, err
	}

	refreshToken := database.NewRefreshToken(userID, refreshTokenStr, expiresAt)

	if err = p.userRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		p.log.Error("create refresh token in db", zap.String("userID", userID), zap.Error(err))
		return RefreshToken{}, err
	}

	return RefreshToken{
		Token:     refreshTokenStr,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshAccessToken validates refresh token and returns new access token
func (p *Provider) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (string, error) {
	// get refresh token from database
	refreshToken, err := p.userRepo.GetRefreshTokenByToken(ctx, refreshTokenStr)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return "", ErrRefreshTokenNotFound
		}
		p.log.Error("get refresh token", zap.Error(err))
		return "", err
	}

	// check if token is expired
	if refreshToken.IsExpired() {
		// delete expired token
		go func() {
			bCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
			defer cancel()

			if delErr := p.userRepo.DeleteRefreshToken(bCtx, refreshTokenStr); delErr != nil {
				p.log.Error("delete expired refresh token", zap.Error(delErr))
			}
		}()

		return "", ErrRefreshTokenExpired
	}

	// get user
	user, err := p.userRepo.GetUserByID(ctx, refreshToken.UserID)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			p.log.Error("get user by id", zap.String("userID", refreshToken.UserID), zap.Error(err))
			return "", err
		}

		// user was deleted, clean up token
		go func() {
			bCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
			defer cancel()

			if delErr := p.userRepo.DeleteRefreshToken(bCtx, refreshTokenStr); delErr != nil {
				p.log.Error("delete refresh token for deleted user", zap.Error(delErr))
			}
		}()

		return "", ErrRefreshTokenNotFound
	}

	// create new access token
	claims := p.auth.CreateUserClaims(mapDBUserToUser(user))
	accessToken, err := p.auth.GenerateToken(claims)
	if err != nil {
		p.log.Error("generate access token", zap.String("userID", user.ID), zap.Error(err))
		return "", err
	}

	return accessToken, nil
}

// ValidateAccessToken validates access token and returns claims from it
func (p *Provider) ValidateAccessToken(tokenStr string) (auth.Claims, error) {
	return p.auth.ValidateToken(tokenStr)
}
