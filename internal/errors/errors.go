// Package errors defines the error hierarchy and exit code mapping for marketsurge-agent.
package errors

import (
	"errors"
)

// Exit code constants for different error types.
const (
	ExitSuccess     = 0
	ExitGeneral     = 1
	ExitNotFound    = 2
	ExitAuthFailure = 3
	ExitAPIError    = 4
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

// As implements custom error type matching for MarketSurgeError.
func (e *MarketSurgeError) As(target interface{}) bool {
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

// AuthenticationError indicates an authentication failure (cookie extraction or token expiry).
type AuthenticationError struct {
	MarketSurgeError
}

// As implements custom error type matching for AuthenticationError.
func (e *AuthenticationError) As(target interface{}) bool {
	if _, ok := target.(**AuthenticationError); ok {
		return true
	}
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

// CookieExtractionError indicates failure to extract browser cookies.
// It is a subtype of AuthenticationError.
type CookieExtractionError struct {
	MarketSurgeError
	Browser string // Name of the browser that failed extraction
}

// As implements custom error type matching for CookieExtractionError.
func (e *CookieExtractionError) As(target interface{}) bool {
	if _, ok := target.(**CookieExtractionError); ok {
		return true
	}
	if _, ok := target.(**AuthenticationError); ok {
		return true
	}
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

// TokenExpiredError indicates that the JWT token has expired or been rejected.
// It is a subtype of AuthenticationError.
type TokenExpiredError struct {
	MarketSurgeError
	StatusCode int // HTTP status code that triggered this error (usually 401)
}

// As implements custom error type matching for TokenExpiredError.
func (e *TokenExpiredError) As(target interface{}) bool {
	if _, ok := target.(**TokenExpiredError); ok {
		return true
	}
	if _, ok := target.(**AuthenticationError); ok {
		return true
	}
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

// APIError indicates that the GraphQL API returned errors.
type APIError struct {
	MarketSurgeError
}

// As implements custom error type matching for APIError.
func (e *APIError) As(target interface{}) bool {
	if _, ok := target.(**APIError); ok {
		return true
	}
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

// SymbolNotFoundError indicates that the API returned empty marketData for a requested symbol.
// It is a subtype of APIError.
type SymbolNotFoundError struct {
	MarketSurgeError
	Symbol string // The ticker symbol that was not found
}

// As implements custom error type matching for SymbolNotFoundError.
func (e *SymbolNotFoundError) As(target interface{}) bool {
	if _, ok := target.(**SymbolNotFoundError); ok {
		return true
	}
	if _, ok := target.(**APIError); ok {
		return true
	}
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

// HTTPError indicates that an HTTP request returned a non-2xx error status code.
type HTTPError struct {
	MarketSurgeError
	StatusCode int    // HTTP status code from the response
	Body       string // Raw response body text
}

// As implements custom error type matching for HTTPError.
func (e *HTTPError) As(target interface{}) bool {
	if _, ok := target.(**HTTPError); ok {
		return true
	}
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

// ValidationError indicates that input validation failed.
type ValidationError struct {
	MarketSurgeError
}

// As implements custom error type matching for ValidationError.
func (e *ValidationError) As(target interface{}) bool {
	if _, ok := target.(**ValidationError); ok {
		return true
	}
	if _, ok := target.(**MarketSurgeError); ok {
		return true
	}
	return false
}

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
// It checks error types from most specific to least specific using errors.As().
// Returns ExitSuccess (0) if err is nil, otherwise returns the appropriate exit code.
func ExitCodeFor(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Check most specific types first
	var symbolNotFound *SymbolNotFoundError
	if errors.As(err, &symbolNotFound) {
		return ExitNotFound
	}

	var cookieExtraction *CookieExtractionError
	if errors.As(err, &cookieExtraction) {
		return ExitAuthFailure
	}

	var tokenExpired *TokenExpiredError
	if errors.As(err, &tokenExpired) {
		return ExitAuthFailure
	}

	var authentication *AuthenticationError
	if errors.As(err, &authentication) {
		return ExitAuthFailure
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return ExitAPIError
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return ExitAPIError
	}

	var validation *ValidationError
	if errors.As(err, &validation) {
		return ExitGeneral
	}

	var marketSurge *MarketSurgeError
	if errors.As(err, &marketSurge) {
		return ExitGeneral
	}

	// Default to general error for unknown error types
	return ExitGeneral
}
