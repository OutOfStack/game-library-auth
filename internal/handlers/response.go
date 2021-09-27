package handlers

// ErrResp describes error response.
type ErrResp struct {
	Error string `json:"error"`
}

// TokenResp describes response with JWT
type TokenResp struct {
	Token string
}
