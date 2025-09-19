package handlers

const (
	internalErrorMsg           = "Internal error"
	validationErrorMsg         = "Validation error"
	authErrorMsg               = "Incorrect username or password"
	invalidAuthTokenMsg        = "Invalid or missing authorization token"
	invalidOrExpiredVrfCodeMsg = "Invalid or expired verification code"
)

// SignInReq represents user sign in request
type SignInReq struct {
	Username string `json:"username"`
	Login    string `json:"login"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

// TokenResp represents response with JWT
type TokenResp struct {
	AccessToken string `json:"accessToken"`
}

// SignUpReq represents user sign up request
type SignUpReq struct {
	Username        string `json:"username" validate:"required,fieldexcludes=@"`
	DisplayName     string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"omitempty,email"`
	Password        string `json:"password" validate:"required,min=8,max=64"`
	ConfirmPassword string `json:"confirmPassword" validate:"eqfield=Password"`
	IsPublisher     bool   `json:"isPublisher"`
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

// VerifyEmailReq represents email verification request with 6-digit code
type VerifyEmailReq struct {
	Code string `json:"code" validate:"required,len=6"`
}

// GoogleOAuthRequest represents Google OAuth request
type GoogleOAuthRequest struct {
	IDToken string `json:"idToken" validate:"required"`
}

type googleIDTokenClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}
