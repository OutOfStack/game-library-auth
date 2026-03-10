package githubapi

// UserInfo represents GitHub user information
type UserInfo struct {
	ID    string
	Login string
	Email string
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
