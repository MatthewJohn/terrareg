package csrf

import "errors"

var (
	// ErrNoSession indicates no session is present to validate CSRF token
	ErrNoSession = errors.New("no session is present to check CSRF token")

	// ErrMissingToken indicates CSRF token was not provided
	ErrMissingToken = errors.New("CSRF token is missing")

	// ErrInvalidToken indicates CSRF token is incorrect
	ErrInvalidToken = errors.New("CSRF token is incorrect")
)