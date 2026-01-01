package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestJWKVerifier_FetchJWKS(t *testing.T) {
	// Create a mock HTTP server that serves JWKS
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/jwks" {
			t.Errorf("Expected path /api/auth/jwks, got %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// Return a mock JWKS response (Ed25519 format)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"keys": [
				{
					"kid": "test-key-1",
					"kty": "OKP",
					"alg": "EdDSA",
					"crv": "Ed25519",
					"x": "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo"
				}
			]
		}`))
	}))
	defer mockServer.Close()

	// Create verifier pointing to mock server
	verifier := NewJWKVerifier(mockServer.URL)

	// Test fetching JWKS
	err := verifier.Initialize()
	if err != nil {
		t.Fatalf("Failed to fetch JWKS: %v", err)
	}

	// Verify that keys were fetched
	if verifier.set == nil {
		t.Error("JWK set should not be nil after initialization")
	}
}

func TestJWKVerifier_Middleware_MissingAuth(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"keys": []}`))
	}))
	defer mockServer.Close()

	verifier := NewJWKVerifier(mockServer.URL)
	verifier.Initialize()

	// Create a simple handler
	handler := verifier.JWTMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Test without Authorization header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Error("Expected error for missing Authorization header")
	}

	// Check that it's an HTTP error with 401 status
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusUnauthorized {
		t.Errorf("Expected HTTP 401 error, got: %v", err)
	}
}

func TestJWKVerifier_Middleware_InvalidFormat(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"keys": []}`))
	}))
	defer mockServer.Close()

	verifier := NewJWKVerifier(mockServer.URL)
	verifier.Initialize()

	handler := verifier.JWTMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Test with invalid Authorization header format
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Error("Expected error for invalid Authorization header format")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusUnauthorized {
		t.Errorf("Expected HTTP 401 error, got: %v", err)
	}
}

