package errors

import (
	"errors"
	"testing"
)

// TestExitCodeForSymbolNotFound verifies SymbolNotFoundError maps to ExitNotFound.
func TestExitCodeForSymbolNotFound(t *testing.T) {
	t.Parallel()
	err := NewSymbolNotFoundError("symbol INVALID not found", nil, "INVALID")
	code := ExitCodeFor(err)
	if code != ExitNotFound {
		t.Errorf("ExitCodeFor(SymbolNotFoundError) = %d, want %d", code, ExitNotFound)
	}
}

// TestExitCodeForSymbolNotFoundWrapped verifies wrapped SymbolNotFoundError still maps correctly.
func TestExitCodeForSymbolNotFoundWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("api returned empty data")
	err := NewSymbolNotFoundError("symbol INVALID not found", underlying, "INVALID")
	code := ExitCodeFor(err)
	if code != ExitNotFound {
		t.Errorf("ExitCodeFor(wrapped SymbolNotFoundError) = %d, want %d", code, ExitNotFound)
	}
}

// TestExitCodeForCookieExtractionError verifies CookieExtractionError maps to ExitAuthFailure.
func TestExitCodeForCookieExtractionError(t *testing.T) {
	t.Parallel()
	err := NewCookieExtractionError("failed to extract cookies", nil, "Firefox")
	code := ExitCodeFor(err)
	if code != ExitAuthFailure {
		t.Errorf("ExitCodeFor(CookieExtractionError) = %d, want %d", code, ExitAuthFailure)
	}
}

// TestExitCodeForCookieExtractionErrorWrapped verifies wrapped CookieExtractionError maps correctly.
func TestExitCodeForCookieExtractionErrorWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("rookiepy failed")
	err := NewCookieExtractionError("failed to extract cookies", underlying, "Chrome")
	code := ExitCodeFor(err)
	if code != ExitAuthFailure {
		t.Errorf("ExitCodeFor(wrapped CookieExtractionError) = %d, want %d", code, ExitAuthFailure)
	}
}

// TestExitCodeForTokenExpiredError verifies TokenExpiredError maps to ExitAuthFailure.
func TestExitCodeForTokenExpiredError(t *testing.T) {
	t.Parallel()
	err := NewTokenExpiredError("token expired", nil, 401)
	code := ExitCodeFor(err)
	if code != ExitAuthFailure {
		t.Errorf("ExitCodeFor(TokenExpiredError) = %d, want %d", code, ExitAuthFailure)
	}
}

// TestExitCodeForTokenExpiredErrorWrapped verifies wrapped TokenExpiredError maps correctly.
func TestExitCodeForTokenExpiredErrorWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("http 401 unauthorized")
	err := NewTokenExpiredError("token expired", underlying, 401)
	code := ExitCodeFor(err)
	if code != ExitAuthFailure {
		t.Errorf("ExitCodeFor(wrapped TokenExpiredError) = %d, want %d", code, ExitAuthFailure)
	}
}

// TestExitCodeForAuthenticationError verifies AuthenticationError maps to ExitAuthFailure.
func TestExitCodeForAuthenticationError(t *testing.T) {
	t.Parallel()
	err := NewAuthenticationError("authentication failed", nil)
	code := ExitCodeFor(err)
	if code != ExitAuthFailure {
		t.Errorf("ExitCodeFor(AuthenticationError) = %d, want %d", code, ExitAuthFailure)
	}
}

// TestExitCodeForAuthenticationErrorWrapped verifies wrapped AuthenticationError maps correctly.
func TestExitCodeForAuthenticationErrorWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("invalid credentials")
	err := NewAuthenticationError("authentication failed", underlying)
	code := ExitCodeFor(err)
	if code != ExitAuthFailure {
		t.Errorf("ExitCodeFor(wrapped AuthenticationError) = %d, want %d", code, ExitAuthFailure)
	}
}

// TestExitCodeForAPIError verifies APIError maps to ExitAPIError.
func TestExitCodeForAPIError(t *testing.T) {
	t.Parallel()
	err := NewAPIError("graphql api error", nil)
	code := ExitCodeFor(err)
	if code != ExitAPIError {
		t.Errorf("ExitCodeFor(APIError) = %d, want %d", code, ExitAPIError)
	}
}

// TestExitCodeForAPIErrorWrapped verifies wrapped APIError maps correctly.
func TestExitCodeForAPIErrorWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("graphql returned errors")
	err := NewAPIError("graphql api error", underlying)
	code := ExitCodeFor(err)
	if code != ExitAPIError {
		t.Errorf("ExitCodeFor(wrapped APIError) = %d, want %d", code, ExitAPIError)
	}
}

// TestExitCodeForHTTPError verifies HTTPError maps to ExitAPIError.
func TestExitCodeForHTTPError(t *testing.T) {
	t.Parallel()
	err := NewHTTPError("http 500 error", nil, 500, "Internal Server Error")
	code := ExitCodeFor(err)
	if code != ExitAPIError {
		t.Errorf("ExitCodeFor(HTTPError) = %d, want %d", code, ExitAPIError)
	}
}

// TestExitCodeForHTTPErrorWrapped verifies wrapped HTTPError maps correctly.
func TestExitCodeForHTTPErrorWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("connection failed")
	err := NewHTTPError("http 502 error", underlying, 502, "Bad Gateway")
	code := ExitCodeFor(err)
	if code != ExitAPIError {
		t.Errorf("ExitCodeFor(wrapped HTTPError) = %d, want %d", code, ExitAPIError)
	}
}

// TestExitCodeForValidationError verifies ValidationError maps to ExitGeneral.
func TestExitCodeForValidationError(t *testing.T) {
	t.Parallel()
	err := NewValidationError("invalid input", nil)
	code := ExitCodeFor(err)
	if code != ExitGeneral {
		t.Errorf("ExitCodeFor(ValidationError) = %d, want %d", code, ExitGeneral)
	}
}

// TestExitCodeForValidationErrorWrapped verifies wrapped ValidationError maps correctly.
func TestExitCodeForValidationErrorWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("symbol is empty")
	err := NewValidationError("invalid input", underlying)
	code := ExitCodeFor(err)
	if code != ExitGeneral {
		t.Errorf("ExitCodeFor(wrapped ValidationError) = %d, want %d", code, ExitGeneral)
	}
}

// TestExitCodeForMarketSurgeError verifies MarketSurgeError maps to ExitGeneral.
func TestExitCodeForMarketSurgeError(t *testing.T) {
	t.Parallel()
	err := NewMarketSurgeError("generic error", nil)
	code := ExitCodeFor(err)
	if code != ExitGeneral {
		t.Errorf("ExitCodeFor(MarketSurgeError) = %d, want %d", code, ExitGeneral)
	}
}

// TestExitCodeForMarketSurgeErrorWrapped verifies wrapped MarketSurgeError maps correctly.
func TestExitCodeForMarketSurgeErrorWrapped(t *testing.T) {
	t.Parallel()
	underlying := errors.New("something went wrong")
	err := NewMarketSurgeError("generic error", underlying)
	code := ExitCodeFor(err)
	if code != ExitGeneral {
		t.Errorf("ExitCodeFor(wrapped MarketSurgeError) = %d, want %d", code, ExitGeneral)
	}
}

// TestExitCodeForStandardError verifies standard errors.New() maps to ExitGeneral.
func TestExitCodeForStandardError(t *testing.T) {
	t.Parallel()
	err := errors.New("standard error")
	code := ExitCodeFor(err)
	if code != ExitGeneral {
		t.Errorf("ExitCodeFor(standard error) = %d, want %d", code, ExitGeneral)
	}
}

// TestExitCodeForNilError verifies nil error maps to ExitSuccess.
func TestExitCodeForNilError(t *testing.T) {
	t.Parallel()
	code := ExitCodeFor(nil)
	if code != ExitSuccess {
		t.Errorf("ExitCodeFor(nil) = %d, want %d", code, ExitSuccess)
	}
}

// TestErrorChainPreservation verifies that error chains are preserved through wrapping.
func TestErrorChainPreservation(t *testing.T) {
	t.Parallel()
	underlying := errors.New("root cause")
	err := NewCookieExtractionError("cookie extraction failed", underlying, "Firefox")

	// Verify the error chain is preserved
	if !errors.Is(err, underlying) {
		t.Error("error chain not preserved: errors.Is(err, underlying) returned false")
	}

	// Verify Unwrap works
	if !errors.Is(err.Unwrap(), underlying) {
		t.Error("Unwrap() did not return the underlying error")
	}
}

// TestErrorChainTraversal verifies that errors.As() can traverse the error chain.
func TestErrorChainTraversal(t *testing.T) {
	t.Parallel()
	underlying := errors.New("root cause")
	err := NewCookieExtractionError("cookie extraction failed", underlying, "Firefox")

	// Verify errors.As() can find CookieExtractionError
	var cookieErr *CookieExtractionError
	if !errors.As(err, &cookieErr) {
		t.Error("errors.As() failed to find CookieExtractionError in chain")
	}

	// Verify errors.As() can find AuthenticationError (parent type)
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Error("errors.As() failed to find AuthenticationError in chain")
	}

	// Verify errors.As() can find MarketSurgeError (grandparent type)
	var marketErr *MarketSurgeError
	if !errors.As(err, &marketErr) {
		t.Error("errors.As() failed to find MarketSurgeError in chain")
	}
}

// TestSymbolNotFoundErrorAttributes verifies SymbolNotFoundError stores the symbol.
func TestSymbolNotFoundErrorAttributes(t *testing.T) {
	t.Parallel()
	symbol := "INVALID"
	err := NewSymbolNotFoundError("symbol not found", nil, symbol)
	if err.Symbol != symbol {
		t.Errorf("SymbolNotFoundError.Symbol = %q, want %q", err.Symbol, symbol)
	}
}

// TestCookieExtractionErrorAttributes verifies CookieExtractionError stores the browser.
func TestCookieExtractionErrorAttributes(t *testing.T) {
	t.Parallel()
	browser := "Firefox"
	err := NewCookieExtractionError("cookie extraction failed", nil, browser)
	if err.Browser != browser {
		t.Errorf("CookieExtractionError.Browser = %q, want %q", err.Browser, browser)
	}
}

// TestTokenExpiredErrorAttributes verifies TokenExpiredError stores the status code.
func TestTokenExpiredErrorAttributes(t *testing.T) {
	t.Parallel()
	statusCode := 401
	err := NewTokenExpiredError("token expired", nil, statusCode)
	if err.StatusCode != statusCode {
		t.Errorf("TokenExpiredError.StatusCode = %d, want %d", err.StatusCode, statusCode)
	}
}

// TestHTTPErrorAttributes verifies HTTPError stores status code and body.
func TestHTTPErrorAttributes(t *testing.T) {
	t.Parallel()
	statusCode := 500
	body := "Internal Server Error"
	err := NewHTTPError("http error", nil, statusCode, body)
	if err.StatusCode != statusCode {
		t.Errorf("HTTPError.StatusCode = %d, want %d", err.StatusCode, statusCode)
	}
	if err.Body != body {
		t.Errorf("HTTPError.Body = %q, want %q", err.Body, body)
	}
}

// TestErrorMessagePreservation verifies that error messages are preserved.
func TestErrorMessagePreservation(t *testing.T) {
	t.Parallel()
	message := "test error message"
	err := NewMarketSurgeError(message, nil)
	if err.Error() != message {
		t.Errorf("Error() = %q, want %q", err.Error(), message)
	}
}

// TestCookieExtractionIsAuthenticationError verifies CookieExtractionError is also AuthenticationError.
func TestCookieExtractionIsAuthenticationError(t *testing.T) {
	t.Parallel()
	err := NewCookieExtractionError("cookie extraction failed", nil, "Firefox")
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Error("CookieExtractionError is not recognized as AuthenticationError")
	}
}

// TestTokenExpiredIsAuthenticationError verifies TokenExpiredError is also AuthenticationError.
func TestTokenExpiredIsAuthenticationError(t *testing.T) {
	t.Parallel()
	err := NewTokenExpiredError("token expired", nil, 401)
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Error("TokenExpiredError is not recognized as AuthenticationError")
	}
}

// TestSymbolNotFoundIsAPIError verifies SymbolNotFoundError is also APIError.
func TestSymbolNotFoundIsAPIError(t *testing.T) {
	t.Parallel()
	err := NewSymbolNotFoundError("symbol not found", nil, "INVALID")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Error("SymbolNotFoundError is not recognized as APIError")
	}
}

// TestExitCodePriority verifies that more specific error types take priority.
// SymbolNotFoundError should map to ExitNotFound, not ExitAPIError.
func TestExitCodePriority(t *testing.T) {
	t.Parallel()
	err := NewSymbolNotFoundError("symbol not found", nil, "INVALID")
	code := ExitCodeFor(err)
	if code != ExitNotFound {
		t.Errorf("ExitCodeFor(SymbolNotFoundError) = %d, want %d (ExitNotFound takes priority over ExitAPIError)", code, ExitNotFound)
	}
}
