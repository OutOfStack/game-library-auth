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
	Name            string `json:"name" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password"`
	IsPublisher     bool   `json:"isPublisher"`
	AvatarURL       string `json:"avatarUrl" validate:"len=0|url"`
}

// SignUpResp represents sign up response
type SignUpResp struct {
	ID uuid.UUID `json:"id"`
}

// UpdateProfileReq represents update profile request
type UpdateProfileReq struct {
	UserID             string  `json:"userId" validate:"uuid4,required"`
	Name               *string `json:"name"`
	AvatarURL          *string `json:"avatarUrl" validate:"omitempty,len=0|url"`
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
