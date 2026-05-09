// Package errors defines the error hierarchy and exit code mapping for marketsurge-agent.
package errors

import (
	"errors"
)

// Exit code ranges:
//   0:     Success
//   1-9:   Reserved (general errors)
//   30+:   Domain-specific errors (MarketSurge)
const (
	ExitSuccess     = 0
	ExitGeneral     = 30
	ExitNotFound    = 31
	ExitAuthFailure = 32
	ExitAPIError    = 33
)

// MarketSurgeError is the base error type for all marketsurge-agent errors.
// It wraps an underlying error to preserve the error chain.
type MarketSurgeError struct {
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *MarketSurgeError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error, enabling error chain traversal.
func (e *MarketSurgeError) Unwrap() error {
	return e.Cause
}

// ErrorCode returns the error classification code for JSON responses.
func (e *MarketSurgeError) ErrorCode() string { return "GENERAL_ERROR" }

// ExitCode returns the process exit code for this error type.
func (e *MarketSurgeError) ExitCode() int { return ExitGeneral }

// AuthenticationError indicates an authentication failure (cookie extraction or token expiry).
type AuthenticationError struct {
	MarketSurgeError
}

// ErrorCode returns the error classification code for JSON responses.
func (e *AuthenticationError) ErrorCode() string { return "AUTH_FAILED" }

// ExitCode returns the process exit code for this error type.
func (e *AuthenticationError) ExitCode() int { return ExitAuthFailure }

// CookieExtractionError indicates failure to extract browser cookies.
// It is a subtype of AuthenticationError.
type CookieExtractionError struct {
	MarketSurgeError
	Browser string // Name of the browser that failed extraction
}

// ErrorCode returns the error classification code for JSON responses.
func (e *CookieExtractionError) ErrorCode() string { return "AUTH_FAILED" }

// ExitCode returns the process exit code for this error type.
func (e *CookieExtractionError) ExitCode() int { return ExitAuthFailure }

// TokenExpiredError indicates that the JWT token has expired or been rejected.
// It is a subtype of AuthenticationError.
type TokenExpiredError struct {
	MarketSurgeError
	StatusCode int // HTTP status code that triggered this error (usually 401)
}

// ErrorCode returns the error classification code for JSON responses.
func (e *TokenExpiredError) ErrorCode() string { return "AUTH_FAILED" }

// ExitCode returns the process exit code for this error type.
func (e *TokenExpiredError) ExitCode() int { return ExitAuthFailure }

// APIError indicates that the GraphQL API returned errors.
type APIError struct {
	MarketSurgeError
}

// ErrorCode returns the error classification code for JSON responses.
func (e *APIError) ErrorCode() string { return "API_ERROR" }

// ExitCode returns the process exit code for this error type.
func (e *APIError) ExitCode() int { return ExitAPIError }

// SymbolNotFoundError indicates that the API returned empty marketData for a requested symbol.
// It is a subtype of APIError.
type SymbolNotFoundError struct {
	MarketSurgeError
	Symbol string // The ticker symbol that was not found
}

// ErrorCode returns the error classification code for JSON responses.
func (e *SymbolNotFoundError) ErrorCode() string { return "SYMBOL_NOT_FOUND" }

// ExitCode returns the process exit code for this error type.
func (e *SymbolNotFoundError) ExitCode() int { return ExitNotFound }

// HTTPError indicates that an HTTP request returned a non-2xx error status code.
type HTTPError struct {
	MarketSurgeError
	StatusCode int    // HTTP status code from the response
	Body       string // Raw response body text
}

// ErrorCode returns the error classification code for JSON responses.
func (e *HTTPError) ErrorCode() string { return "HTTP_ERROR" }

// ExitCode returns the process exit code for this error type.
func (e *HTTPError) ExitCode() int { return ExitAPIError }

// ValidationError indicates that input validation failed.
type ValidationError struct {
	MarketSurgeError
}

// ErrorCode returns the error classification code for JSON responses.
func (e *ValidationError) ErrorCode() string { return "VALIDATION_ERROR" }

// ExitCode returns the process exit code for this error type.
func (e *ValidationError) ExitCode() int { return ExitGeneral }

// NewMarketSurgeError creates a new MarketSurgeError wrapping the given cause.
func NewMarketSurgeError(message string, cause error) *MarketSurgeError {
	return &MarketSurgeError{
		Message: message,
		Cause:   cause,
	}
}

// NewAuthenticationError creates a new AuthenticationError wrapping the given cause.
func NewAuthenticationError(message string, cause error) *AuthenticationError {
	return &AuthenticationError{
		MarketSurgeError: MarketSurgeError{
			Message: message,
			Cause:   cause,
		},
	}
}

// NewCookieExtractionError creates a new CookieExtractionError wrapping the given cause.
func NewCookieExtractionError(message string, cause error, browser string) *CookieExtractionError {
	return &CookieExtractionError{
		MarketSurgeError: MarketSurgeError{
			Message: message,
			Cause:   cause,
		},
		Browser: browser,
	}
}

// NewTokenExpiredError creates a new TokenExpiredError wrapping the given cause.
func NewTokenExpiredError(message string, cause error, statusCode int) *TokenExpiredError {
	return &TokenExpiredError{
		MarketSurgeError: MarketSurgeError{
			Message: message,
			Cause:   cause,
		},
		StatusCode: statusCode,
	}
}

// NewAPIError creates a new APIError wrapping the given cause.
func NewAPIError(message string, cause error) *APIError {
	return &APIError{
		MarketSurgeError: MarketSurgeError{
			Message: message,
			Cause:   cause,
		},
	}
}

// NewSymbolNotFoundError creates a new SymbolNotFoundError wrapping the given cause.
func NewSymbolNotFoundError(message string, cause error, symbol string) *SymbolNotFoundError {
	return &SymbolNotFoundError{
		MarketSurgeError: MarketSurgeError{
			Message: message,
			Cause:   cause,
		},
		Symbol: symbol,
	}
}

// NewHTTPError creates a new HTTPError wrapping the given cause.
func NewHTTPError(message string, cause error, statusCode int, body string) *HTTPError {
	return &HTTPError{
		MarketSurgeError: MarketSurgeError{
			Message: message,
			Cause:   cause,
		},
		StatusCode: statusCode,
		Body:       body,
	}
}

// NewValidationError creates a new ValidationError wrapping the given cause.
func NewValidationError(message string, cause error) *ValidationError {
	return &ValidationError{
		MarketSurgeError: MarketSurgeError{
			Message: message,
			Cause:   cause,
		},
	}
}

// ExitCodeFor determines the appropriate exit code for the given error.
// Each error type implements ExitCode() to declare its own exit code.
// Returns ExitSuccess (0) if err is nil, otherwise returns the appropriate exit code.
func ExitCodeFor(err error) int {
	if err == nil {
		return ExitSuccess
	}

	type exitCoder interface {
		error
		ExitCode() int
	}

	if coder, ok := errors.AsType[exitCoder](err); ok {
		return coder.ExitCode()
	}

	return ExitGeneral
}
