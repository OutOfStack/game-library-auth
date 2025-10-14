package resendapi

// SendEmailVerificationRequest represents email verification request
type SendEmailVerificationRequest struct {
	Email            string
	Username         string
	VerificationCode string
	UnsubscribeToken string
}

// templateData represents data passed to email templates
type templateData struct {
	Email             string
	Username          string
	VerificationCode  string
	UnsubscribeToken  string
	BaseURL           string
	UnsubscribeURL    string
	ContactEmail      string
	PrivacyPolicyURL  string
	TermsOfServiceURL string
	CurrentYear       int
}
