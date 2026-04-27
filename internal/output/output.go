// Package output provides JSON envelope types and writers for marketsurge-agent responses.
package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

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
// It maps the error type to an appropriate error code string using errors.AsType().
// The response is formatted as a JSON envelope with error details.
func WriteError(w io.Writer, err error) error {
	code := errorCode(err)
	message := err.Error()

	// Extract details from specific error types
	var details string
	if symbolNotFound, ok := errors.AsType[*mserr.SymbolNotFoundError](err); ok {
		details = "symbol: " + symbolNotFound.Symbol
	}

	if httpErr, ok := errors.AsType[*mserr.HTTPError](err); ok {
		details = httpErrorDetails(httpErr)
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

const maxHTTPErrorBodyDetailLength = 500

var sensitiveHTTPBodyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`),
	regexp.MustCompile(`(?i)Bearer\s+\S+`),
	regexp.MustCompile(`\b[A-Za-z_][A-Za-z0-9_-]*=[^\s;,]+`),
}

func httpErrorDetails(err *mserr.HTTPError) string {
	details := fmt.Sprintf("status: %d", err.StatusCode)
	if err.Body == "" {
		return details
	}

	body := truncateHTTPErrorBody(redactHTTPErrorBody(err.Body))

	return fmt.Sprintf("%s; body: %s", details, body)
}

func redactHTTPErrorBody(body string) string {
	for _, pattern := range sensitiveHTTPBodyPatterns {
		body = pattern.ReplaceAllString(body, "[REDACTED]")
	}
	return body
}

func truncateHTTPErrorBody(body string) string {
	runes := []rune(body)
	if len(runes) <= maxHTTPErrorBodyDetailLength {
		return body
	}

	return string(runes[:maxHTTPErrorBodyDetailLength]) + "..."
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
// Each error type implements ErrorCode() to declare its own classification code.
func errorCode(err error) string {
	type errorCoder interface {
		error
		ErrorCode() string
	}

	if coder, ok := errors.AsType[errorCoder](err); ok {
		return coder.ErrorCode()
	}

	return "GENERAL_ERROR"
}
