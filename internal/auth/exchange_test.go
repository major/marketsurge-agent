package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/marketsurge-agent/internal/errors"
	marketsurge "github.com/major/marketsurge-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type clientResponse = marketsurge.ClientInfoResponse

// validJWTResponse returns a clientResponse representing a successful login.
func validJWTResponse() clientResponse {
	return clientResponse{
		IsLoggedIn: true,
		JWT:        "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test.signature",
		GivenName:  "Test",
		FamilyName: "User",
	}
}

func validSessionCookies() []*http.Cookie {
	return []*http.Cookie{
		{Name: "session", Value: "abc123"},
	}
}

func TestExchangeJWT_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(validJWTResponse()))
	}))
	defer server.Close()

	origURL := exchangeURL
	exchangeURL = server.URL
	t.Cleanup(func() { exchangeURL = origURL })

	jwt, err := ExchangeJWT(context.Background(), validSessionCookies())

	require.NoError(t, err)
	assert.Equal(t, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test.signature", jwt)
}

func TestExchangeJWT_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	origURL := exchangeURL
	exchangeURL = server.URL
	t.Cleanup(func() { exchangeURL = origURL })

	jwt, err := ExchangeJWT(context.Background(), validSessionCookies())

	assert.Empty(t, jwt)
	require.Error(t, err)

	var authErr *errors.AuthenticationError
	require.ErrorAs(t, err, &authErr)
	assert.Contains(t, authErr.Message, "JWT exchange failed: HTTP 401")
}

func TestExchangeJWT_NoJWT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := clientResponse{
			IsLoggedIn: true,
			JWT:        "",
			GivenName:  "Test",
			FamilyName: "User",
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	origURL := exchangeURL
	exchangeURL = server.URL
	t.Cleanup(func() { exchangeURL = origURL })

	jwt, err := ExchangeJWT(context.Background(), validSessionCookies())

	assert.Empty(t, jwt)
	require.Error(t, err)

	var authErr *errors.AuthenticationError
	require.ErrorAs(t, err, &authErr)
	assert.Contains(t, authErr.Message, "no JWT found in response")
}

func TestExchangeJWT_NotLoggedIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := clientResponse{
			IsLoggedIn: false,
			JWT:        "",
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	origURL := exchangeURL
	exchangeURL = server.URL
	t.Cleanup(func() { exchangeURL = origURL })

	jwt, err := ExchangeJWT(context.Background(), validSessionCookies())

	assert.Empty(t, jwt)
	require.Error(t, err)

	var authErr *errors.AuthenticationError
	require.ErrorAs(t, err, &authErr)
	assert.Contains(t, authErr.Message, "not logged in")
}

func TestExchangeJWT_Cookies(t *testing.T) {
	var receivedCookies []*http.Cookie

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCookies = r.Cookies()
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(validJWTResponse()))
	}))
	defer server.Close()

	origURL := exchangeURL
	exchangeURL = server.URL
	t.Cleanup(func() { exchangeURL = origURL })

	cookies := []*http.Cookie{
		{Name: "session_id", Value: "sess-abc-123"},
		{Name: "auth_token", Value: "tok-xyz-789"},
		{Name: "preference", Value: "dark_mode"},
	}

	_, err := ExchangeJWT(context.Background(), cookies)
	require.NoError(t, err)

	require.Len(t, receivedCookies, 3)

	cookieMap := make(map[string]string)
	for _, c := range receivedCookies {
		cookieMap[c.Name] = c.Value
	}
	assert.Equal(t, "sess-abc-123", cookieMap["session_id"])
	assert.Equal(t, "tok-xyz-789", cookieMap["auth_token"])
	assert.Equal(t, "dark_mode", cookieMap["preference"])
}

func TestExchangeJWT_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	origURL := exchangeURL
	exchangeURL = server.URL
	t.Cleanup(func() { exchangeURL = origURL })

	jwt, err := ExchangeJWT(context.Background(), validSessionCookies())

	assert.Empty(t, jwt)
	require.Error(t, err)

	var authErr *errors.AuthenticationError
	require.ErrorAs(t, err, &authErr)
	assert.Contains(t, authErr.Message, "JWT exchange request failed")
}
