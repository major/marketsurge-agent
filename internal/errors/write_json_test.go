package errors_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON_AuthenticationError(t *testing.T) {
	t.Parallel()
	err := mserrors.NewAuthenticationError("not logged in", nil)
	var buf bytes.Buffer
	mserrors.WriteJSON(&buf, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "AUTH_FAILED", got["code"])
	assert.Equal(t, "not logged in", got["message"])
}

func TestWriteJSON_ValidationError(t *testing.T) {
	t.Parallel()
	err := mserrors.NewValidationError("invalid argument", nil)
	var buf bytes.Buffer
	mserrors.WriteJSON(&buf, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "VALIDATION_ERROR", got["code"])
	assert.Equal(t, "invalid argument", got["message"])
}

func TestWriteJSON_PlainError(t *testing.T) {
	t.Parallel()
	err := errors.New("something went wrong")
	var buf bytes.Buffer
	mserrors.WriteJSON(&buf, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "GENERAL_ERROR", got["code"])
	assert.Equal(t, "something went wrong", got["message"])
}

func TestWriteJSON_APIError(t *testing.T) {
	t.Parallel()
	err := mserrors.NewAPIError("graphql error", nil)
	var buf bytes.Buffer
	mserrors.WriteJSON(&buf, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "API_ERROR", got["code"])
	assert.Equal(t, "graphql error", got["message"])
}
