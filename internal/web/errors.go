package web

import "fmt"

// FieldError represents error in a struct field
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ValidationError adds information to request error
type ValidationError struct {
	Err    error
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	var fieldsMsg string
	if len(e.Fields) > 0 {
		fieldsMsg = fmt.Sprintf(" - fields: %v", e.Fields)
	}
	return fmt.Sprintf("%s%s", e.Err.Error(), fieldsMsg)
}
