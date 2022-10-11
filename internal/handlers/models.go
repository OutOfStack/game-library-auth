package handlers

import (
	"github.com/google/uuid"
)

// SignInReq represents user sign in request
type SignInReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// TokenResp describes response with JWT
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
}

// SignUpResp represents sign up response
type SignUpResp struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	RoleID      uuid.UUID `json:"roleId"`
	DateCreated string    `json:"dateCreated"`
	DateUpdated string    `json:"dateUpdated"`
}
