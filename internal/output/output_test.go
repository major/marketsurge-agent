package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	mserr "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	data := map[string]any{"symbol": "AAPL", "price": "150.00"}
	metadata := map[string]any{"timestamp": "2026-04-21T10:00:00Z"}

	err := WriteSuccess(buf, data, metadata)
	require.NoError(t, err)

	var envelope Envelope
	err = json.Unmarshal(buf.Bytes(), &envelope)
	require.NoError(t, err)

	assert.Equal(t, data, envelope.Data)
	assert.Equal(t, metadata, envelope.Metadata)
	assert.Nil(t, envelope.Errors)
}

func TestWriteSuccessWithNilMetadata(t *testing.T) {
	buf := &bytes.Buffer{}
	data := map[string]any{"test": "value"}

	err := WriteSuccess(buf, data, nil)
	require.NoError(t, err)

	var envelope Envelope
	err = json.Unmarshal(buf.Bytes(), &envelope)
	require.NoError(t, err)

	assert.Equal(t, data, envelope.Data)
	assert.Nil(t, envelope.Errors)
}

func TestWriteErrorSymbolNotFound(t *testing.T) {
	buf := &bytes.Buffer{}
	err := mserr.NewSymbolNotFoundError("symbol not found", errors.New("api returned empty"), "INVALID")

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "SYMBOL_NOT_FOUND", errorEnvelope.Error.Code)
	assert.Equal(t, "symbol not found", errorEnvelope.Error.Message)
	assert.Contains(t, errorEnvelope.Error.Details, "INVALID")
}

func TestWriteErrorAuthenticationError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := mserr.NewAuthenticationError("auth failed", errors.New("invalid token"))

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "AUTH_FAILED", errorEnvelope.Error.Code)
	assert.Equal(t, "auth failed", errorEnvelope.Error.Message)
}

func TestWriteErrorCookieExtractionError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := mserr.NewCookieExtractionError("cookie extraction failed", errors.New("db error"), "firefox")

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "AUTH_FAILED", errorEnvelope.Error.Code)
}

func TestWriteErrorTokenExpiredError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := mserr.NewTokenExpiredError("token expired", errors.New("401 response"), 401)

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "AUTH_FAILED", errorEnvelope.Error.Code)
}

func TestWriteErrorAPIError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := mserr.NewAPIError("api error", errors.New("graphql error"))

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "API_ERROR", errorEnvelope.Error.Code)
}

func TestWriteErrorHTTPError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := mserr.NewHTTPError("http error", errors.New("500 response"), 500, "Internal Server Error")

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "HTTP_ERROR", errorEnvelope.Error.Code)
}

func TestWriteErrorValidationError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := mserr.NewValidationError("validation failed", errors.New("invalid input"))

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "VALIDATION_ERROR", errorEnvelope.Error.Code)
}

func TestWriteErrorGenericError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := errors.New("unknown error")

	writeErr := WriteError(buf, err)
	require.NoError(t, writeErr)

	var errorEnvelope ErrorEnvelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &errorEnvelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "GENERAL_ERROR", errorEnvelope.Error.Code)
}

func TestWritePartial(t *testing.T) {
	buf := &bytes.Buffer{}
	data := map[string]any{"partial": "data"}
	errs := []string{"error 1", "error 2"}
	metadata := map[string]any{"timestamp": "2026-04-21T10:00:00Z"}

	writeErr := WritePartial(buf, data, errs, metadata)
	require.NoError(t, writeErr)

	var envelope Envelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &envelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, data, envelope.Data)
	assert.Equal(t, errs, envelope.Errors)
	assert.Equal(t, metadata, envelope.Metadata)
}

func TestWritePartialWithEmptyErrors(t *testing.T) {
	buf := &bytes.Buffer{}
	data := map[string]any{"test": "data"}
	metadata := map[string]any{"timestamp": "2026-04-21T10:00:00Z"}

	writeErr := WritePartial(buf, data, []string{}, metadata)
	require.NoError(t, writeErr)

	var envelope Envelope
	unmarshalErr := json.Unmarshal(buf.Bytes(), &envelope)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, data, envelope.Data)
	assert.Nil(t, envelope.Errors)
	assert.Equal(t, metadata, envelope.Metadata)
}

func TestJSONEscapeHTMLDisabled(t *testing.T) {
	buf := &bytes.Buffer{}
	data := map[string]string{"url": "https://example.com?foo=bar&baz=qux"}
	metadata := map[string]any{}

	err := WriteSuccess(buf, data, metadata)
	require.NoError(t, err)

	// Check that & is not escaped to \u0026
	assert.NotContains(t, buf.String(), "\\u0026")
	assert.Contains(t, buf.String(), "&")
}

func TestSymbolMeta(t *testing.T) {
	meta := SymbolMeta("AAPL")

	assert.Equal(t, "AAPL", meta["symbol"])
	assert.NotNil(t, meta["timestamp"])

	// Verify timestamp is a valid RFC3339 string
	timestamp, ok := meta["timestamp"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, timestamp)
}

func TestCatalogMeta(t *testing.T) {
	meta := CatalogMeta("reports", 47, 10, 0)

	assert.Equal(t, "reports", meta["kind"])
	assert.Equal(t, 47, meta["total"])
	assert.Equal(t, 10, meta["limit"])
	assert.Equal(t, 0, meta["offset"])
	assert.NotNil(t, meta["timestamp"])

	// Verify timestamp is a valid RFC3339 string
	timestamp, ok := meta["timestamp"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, timestamp)
}
