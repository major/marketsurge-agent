package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientDefaults(t *testing.T) {
	client := NewClient("jwt-token")

	require.NotNil(t, client.HTTPClient)
	assert.Equal(t, "jwt-token", client.JWT)
	assert.NotEmpty(t, client.Endpoint)
}

func TestExecuteSetsHeadersAndAuthorization(t *testing.T) {
	var captured Request
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		require.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "Bearer jwt-token", r.Header.Get("Authorization"))
		assert.Equal(t, "marketsurge", r.Header.Get("Apollographql-Client-Name"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	})

	raw, err := client.Execute(context.Background(), Request{OperationName: "TestOp", Variables: map[string]any{"value": "x"}, Query: "query TestOp { ok }"})
	require.NoError(t, err)
	assert.Equal(t, "TestOp", captured.OperationName)
	assert.Equal(t, "x", captured.Variables["value"])
	assert.Equal(t, true, getNestedMap(raw, "data")["ok"])
}

func TestExecuteReturnsGraphQLErrorOnHTTP200(t *testing.T) {
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"bad request"}]}`))
	})

	_, err := client.Execute(context.Background(), Request{})
	var apiErr *mserrors.APIError
	require.Error(t, err)
	assert.ErrorAs(t, err, &apiErr)
	assert.Contains(t, err.Error(), "bad request")
}

func TestExecuteReturnsTokenExpiredErrorOn401(t *testing.T) {
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	})

	_, err := client.Execute(context.Background(), Request{})
	var authErr *mserrors.TokenExpiredError
	require.Error(t, err)
	assert.ErrorAs(t, err, &authErr)
}

func TestExecuteReturnsAuthenticationErrorOn403(t *testing.T) {
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	_, err := client.Execute(context.Background(), Request{})
	var authErr *mserrors.AuthenticationError
	require.Error(t, err)
	assert.ErrorAs(t, err, &authErr)
}

func TestExecuteReturnsHTTPErrorOn500(t *testing.T) {
	client := testServerAndClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	_, err := client.Execute(context.Background(), Request{})
	var httpErr *mserrors.HTTPError
	require.Error(t, err)
	assert.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusInternalServerError, httpErr.StatusCode)
	assert.Contains(t, httpErr.Body, "boom")
}
