package handlers

import (
	"github.com/google/uuid"
)

const (
	internalErrorMsg   string = "Internal error"
	validationErrorMsg string = "Validation error"
	authErrorMsg       string = "Incorrect username or password"
)

// SignInReq represents user sign in request
type SignInReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// TokenResp represents response with JWT
type TokenResp struct {
	AccessToken string `json:"accessToken"`
}

// SignUpReq represents user sign up request
type SignUpReq struct {
	Username        string `json:"username" validate:"required"`
	DisplayName     string `json:"name" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password"`
	IsPublisher     bool   `json:"isPublisher"`
}

// SignUpResp represents sign up response
type SignUpResp struct {
	ID uuid.UUID `json:"id"`
}

// UpdateProfileReq represents update profile request
type UpdateProfileReq struct {
	Name               *string `json:"name"`
	Password           *string `json:"password"`
	NewPassword        *string `json:"newPassword" validate:"omitempty,min=8"`
	ConfirmNewPassword *string `json:"confirmNewPassword"`
}

// VerifyTokenReq represents verify JWT request
type VerifyTokenReq struct {
	Token string `json:"token" validate:"jwt"`
}

// VerifyTokenResp represents verify JWT response
type VerifyTokenResp struct {
	Valid bool `json:"valid"`
}

// GoogleOAuthRequest represents Google OAuth request
type GoogleOAuthRequest struct {
	IDToken string `json:"idToken"`
}

type googleIDTokenClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
