package web

// ErrResp is a response in case of an error
type ErrResp struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

// TokenResp describes response with JWT
type TokenResp struct {
	Token string
}
