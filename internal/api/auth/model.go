package auth

const (
	internalErrorMsg           = "Internal error"
	validationErrorMsg         = "Validation error"
	authErrorMsg               = "Incorrect username or password"
	invalidAuthTokenMsg        = "Invalid or missing authorization token"
	invalidOrExpiredVrfCodeMsg = "Invalid or expired verification code"
	invalidRequestMsg          = "Cannot parse request"
	emailAlreadyVerifiedMsg    = "Email is already verified"
	parsingDataErrorMsg        = "Error parsing data"

	refreshTokenCookieName = "refresh_token"
)

type googleIDTokenClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
