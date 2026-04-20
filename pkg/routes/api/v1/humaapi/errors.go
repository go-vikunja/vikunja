package humaapi

import (
	"github.com/danielgtaylor/huma/v2"
)

// vikunjaError is the JSON shape Vikunja's existing error_handler.go emits
// for string-style errors. We preserve it so frontend contracts don't change.
type vikunjaError struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"code,omitempty"`
	Message    string `json:"message"`
}

func (e *vikunjaError) Error() string  { return e.Message }
func (e *vikunjaError) GetStatus() int { return e.StatusCode }

// NewVikunjaError produces an error that serializes to Vikunja's legacy shape
// (`{"message": "..."}` for plain errors; `{"code": X, "message": "..."}` when
// a domain code is supplied). Registered as huma.NewError so every Huma
// handler's error return routes through here.
func NewVikunjaError(status int, msg string, _ ...error) huma.StatusError {
	return &vikunjaError{StatusCode: status, Message: msg}
}

// Install replaces huma.NewError globally. Call once at init.
func Install() {
	huma.NewError = NewVikunjaError
}
