// Package errors defines the error hierarchy and exit code mapping for marketsurge-agent.
package errors

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

var (
	jwtPattern = regexp.MustCompile(`(?i)(?:bearer\s+)?[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)
	secretPairPattern = regexp.MustCompile(`(?i)\b(?:cookie|jwt|token|session|auth)[A-Za-z0-9_-]*=[^\s;]+`)
	browserProfilePathPattern = regexp.MustCompile(`(?i)(?:/home/[^\s:]*\.(?:mozilla|librewolf|waterfox|floorp)[^\s:]*|/Users/[^\s:]*\/(?:Library\/Application Support\/Firefox|\.mozilla)[^\s:]*|[A-Z]:\\Users\\[^\s:]*\\AppData\\Roaming\\Mozilla\\Firefox[^\s:]*)`)
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
	if err == nil {
		return
	}

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
		Message: redactSensitive(err.Error()),
	}
	// Best-effort write; ignore encoder errors (can't report them if stderr is broken).
	enc := json.NewEncoder(w)
	_ = enc.Encode(payload)
}

func redactSensitive(message string) string {
	message = jwtPattern.ReplaceAllString(message, "[REDACTED_JWT]")
	message = secretPairPattern.ReplaceAllString(message, "[REDACTED_SECRET]")
	message = browserProfilePathPattern.ReplaceAllString(message, "[REDACTED_BROWSER_PROFILE]")
	return message
}
