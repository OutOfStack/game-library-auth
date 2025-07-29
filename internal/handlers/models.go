package handlers

import (
	"github.com/google/uuid"
)

const (
	internalErrorMsg   string = "Internal error"
	validationErrorMsg string = "Validation error"
	authErrorMsg       string = "Incorrect username or password"
	invalidAuthToken   string = "Invalid or missing authorization token"

	maxUsernameLen = 32
)

// SignInReq represents user sign in request
type SignInReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

// TokenResp represents response with JWT
type TokenResp struct {
	AccessToken string `json:"accessToken"`
}

// SignUpReq represents user sign up request
type SignUpReq struct {
	Username        string `json:"username" validate:"required"`
	DisplayName     string `json:"name" validate:"required"`
	Password        string `json:"password" validate:"required,min=8,max=64"`
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
	Password           *string `json:"password" validate:"omitempty,min=8,max=64"`
	NewPassword        *string `json:"newPassword" validate:"omitempty,min=8,max=64"`
	ConfirmNewPassword *string `json:"confirmNewPassword" validate:"omitempty,min=8,max=64"`
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
	IDToken string `json:"idToken" validate:"required"`
}

type googleIDTokenClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}
