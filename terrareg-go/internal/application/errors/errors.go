package errors

import "errors"

// Common error types that can be checked with errors.Is()
var (
	// ErrNamespaceNotFound is returned when a namespace is not found
	ErrNamespaceNotFound = errors.New("namespace not found")

	// ErrModuleProviderNotFound is returned when a module provider is not found
	ErrModuleProviderNotFound = errors.New("module provider not found")

	// ErrModuleVersionNotFound is returned when a module version is not found
	ErrModuleVersionNotFound = errors.New("module version not found")

	// ErrModuleVersionNotPublished is returned when a module version is not published
	ErrModuleVersionNotPublished = errors.New("module version is not published")

	// ErrSubmoduleNotFound is returned when a submodule is not found
	ErrSubmoduleNotFound = errors.New("submodule not found")

	// ErrExampleNotFound is returned when an example is not found
	ErrExampleNotFound = errors.New("example not found")

	// ErrFileNotFound is returned when a file is not found
	ErrFileNotFound = errors.New("file not found")

	// ErrNoReadmeContent is returned when no README content is found
	ErrNoReadmeContent = errors.New("no README content found")
)

// WrapNotFound wraps a not-found error with context
func WrapNotFound(base error, details string) error {
	return &NotFoundError{Base: base, Details: details}
}

// NotFoundError represents a not-found error with additional details
type NotFoundError struct {
	Base    error
	Details string
}

func (e *NotFoundError) Error() string {
	if e.Details != "" {
		return e.Base.Error() + ": " + e.Details
	}
	return e.Base.Error()
}

func (e *NotFoundError) Is(target error) bool {
	return e.Base == target
}

// IsNotFound checks if an error is any type of not-found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNamespaceNotFound) ||
		errors.Is(err, ErrModuleProviderNotFound) ||
		errors.Is(err, ErrModuleVersionNotFound) ||
		errors.Is(err, ErrSubmoduleNotFound) ||
		errors.Is(err, ErrExampleNotFound) ||
		errors.Is(err, ErrFileNotFound)
}
