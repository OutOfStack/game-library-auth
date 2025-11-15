package facade

import (
	"context"
	"errors"

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

// RefreshTokens validates refresh token and returns new access and refresh tokens
func (p *Provider) RefreshTokens(ctx context.Context, refreshTokenStr string) (TokenPair, error) {
	var accessToken string
	var newRefreshToken database.RefreshToken
	var deleteToken bool

	txErr := p.userRepo.RunWithTx(ctx, func(txCtx context.Context) error {
		// get refresh token
		refreshToken, err := p.userRepo.GetRefreshTokenByToken(txCtx, refreshTokenStr)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return ErrRefreshTokenNotFound
			}
			p.log.Error("get refresh token", zap.Error(err))
			return err
		}

		// check if token is expired
		if refreshToken.IsExpired() {
			deleteToken = true
			return ErrRefreshTokenExpired
		}

		// get user
		user, err := p.userRepo.GetUserByID(txCtx, refreshToken.UserID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				// user was deleted
				deleteToken = true
				return ErrRefreshTokenNotFound
			}
			p.log.Error("get user by id", zap.String("userID", refreshToken.UserID), zap.Error(err))
			return err
		}

		// generate new access token
		claims := p.auth.CreateUserClaims(mapDBUserToUser(user))
		accessToken, err = p.auth.GenerateToken(claims)
		if err != nil {
			p.log.Error("generate access token", zap.String("userID", user.ID), zap.Error(err))
			return err
		}

		// generate new refresh token
		newRefreshTokenStr, expiresAt, err := p.auth.GenerateRefreshToken()
		if err != nil {
			p.log.Error("generate refresh token", zap.String("userID", user.ID), zap.Error(err))
			return err
		}

		// delete old refresh token
		if err = p.userRepo.DeleteRefreshToken(txCtx, refreshTokenStr); err != nil {
			return err
		}

		// create new refresh token
		newRefreshToken = database.NewRefreshToken(user.ID, newRefreshTokenStr, expiresAt)
		if err = p.userRepo.CreateRefreshToken(txCtx, newRefreshToken); err != nil {
			return err
		}

		return nil
	})
	if txErr != nil {
		// cleanup expired or orphaned tokens only when transaction failed
		if deleteToken {
			if err := p.userRepo.DeleteRefreshToken(ctx, refreshTokenStr); err != nil {
				p.log.Error("delete stale refresh token", zap.Error(err))
			}
		}
		return TokenPair{}, txErr
	}

	return TokenPair{
		AccessToken: accessToken,
		RefreshToken: RefreshToken{
			Token:     newRefreshToken.Token,
			ExpiresAt: newRefreshToken.ExpiresAt,
		},
	}, nil
}

// ValidateAccessToken validates access token and returns claims from it
func (p *Provider) ValidateAccessToken(tokenStr string) (auth.Claims, error) {
	return p.auth.ValidateToken(tokenStr)
}

// RevokeRefreshToken revokes a refresh token by deleting it from the database
func (p *Provider) RevokeRefreshToken(ctx context.Context, refreshTokenStr string) error {
	if refreshTokenStr == "" {
		return nil
	}

	if err := p.userRepo.DeleteRefreshToken(ctx, refreshTokenStr); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// token doesn't exist, nothing to revoke
			return nil
		}
		p.log.Error("revoke refresh token", zap.Error(err))
		return err
	}

	return nil
}
