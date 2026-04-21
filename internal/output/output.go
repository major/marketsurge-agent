// Package output provides JSON envelope types and writers for marketsurge-agent responses.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	mserr "github.com/major/marketsurge-agent/internal/errors"
)

// Envelope is the standard JSON response wrapper for successful operations.
type Envelope struct {
	Data     any            `json:"data"`
	Errors   []string       `json:"errors,omitempty"`
	Metadata map[string]any `json:"metadata"`
}

// ErrorEnvelope is the standard JSON response wrapper for error responses.
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code, message, and optional details.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// WriteSuccess writes a successful response with data and metadata to the writer.
// The response is formatted as a JSON envelope with data and metadata fields.
func WriteSuccess(w io.Writer, data any, metadata map[string]any) error {
	envelope := Envelope{
		Data:     data,
		Metadata: metadata,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(envelope)
}

// WriteError writes an error response to the writer.
// It maps the error type to an appropriate error code string using errors.As().
// The response is formatted as a JSON envelope with error details.
func WriteError(w io.Writer, err error) error {
	code := errorCode(err)
	message := err.Error()

	// Extract details from specific error types
	var details string
	var symbolNotFound *mserr.SymbolNotFoundError
	if errors.As(err, &symbolNotFound) {
		details = "symbol: " + symbolNotFound.Symbol
	}

	var httpErr *mserr.HTTPError
	if errors.As(err, &httpErr) {
		details = fmt.Sprintf("status: %d", httpErr.StatusCode)
	}

	errorEnvelope := ErrorEnvelope{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(errorEnvelope)
}

// WritePartial writes a response with both data and errors (partial success).
// The response is formatted as a JSON envelope with data, errors, and metadata fields.
func WritePartial(w io.Writer, data any, errs []string, metadata map[string]any) error {
	envelope := Envelope{
		Data:     data,
		Errors:   errs,
		Metadata: metadata,
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(envelope)
}

// errorCode maps an error to its corresponding error code string.
// It uses errors.As() to check error types from most specific to least specific.
func errorCode(err error) string {
	// Check most specific types first
	var symbolNotFound *mserr.SymbolNotFoundError
	if errors.As(err, &symbolNotFound) {
		return "SYMBOL_NOT_FOUND"
	}

	var cookieExtraction *mserr.CookieExtractionError
	if errors.As(err, &cookieExtraction) {
		return "AUTH_FAILED"
	}

	var tokenExpired *mserr.TokenExpiredError
	if errors.As(err, &tokenExpired) {
		return "AUTH_FAILED"
	}

	var authentication *mserr.AuthenticationError
	if errors.As(err, &authentication) {
		return "AUTH_FAILED"
	}

	var httpErr *mserr.HTTPError
	if errors.As(err, &httpErr) {
		return "HTTP_ERROR"
	}

	var apiErr *mserr.APIError
	if errors.As(err, &apiErr) {
		return "API_ERROR"
	}

	var validation *mserr.ValidationError
	if errors.As(err, &validation) {
		return "VALIDATION_ERROR"
	}

	var marketSurge *mserr.MarketSurgeError
	if errors.As(err, &marketSurge) {
		return "GENERAL_ERROR"
	}

	// Default to general error for unknown error types
	return "GENERAL_ERROR"
}
