// Package errors defines the error hierarchy and exit code mapping for marketsurge-agent.
package errors

import (
	"encoding/json"
	"errors"
	"io"
)

// errorJSON is the JSON shape for stderr error output.
type errorJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON writes a JSON error object to the given writer.
// Output shape: {"code":"AUTH_FAILED","message":"not logged in -- ..."}
// For typed MarketSurge errors, the code comes from ErrorCode().
// For plain errors, code is "GENERAL_ERROR".
func WriteJSON(w io.Writer, err error) {
	code := "GENERAL_ERROR"

	type errorCoder interface {
		error
		ErrorCode() string
	}

	if coder, ok := errors.AsType[errorCoder](err); ok {
		code = coder.ErrorCode()
	}

	payload := errorJSON{
		Code:    code,
		Message: err.Error(),
	}
	// Best-effort write; ignore encoder errors (can't report them if stderr is broken).
	enc := json.NewEncoder(w)
	_ = enc.Encode(payload)
}
