package githubapi

// UserInfo represents GitHub user information
type UserInfo struct {
	ID            string
	Login         string
	Email         string
	EmailVerified bool
}

// accessTokenResponse represents GitHub access token response
type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// githubUser represents GitHub user API response
type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

// githubEmail represents a GitHub user email from /user/emails API
type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}
