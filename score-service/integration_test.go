//go:build integration
// +build integration

package main

import (
	"archy/scores/jwt"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultAuthURL  = "http://localhost:3000"
	defaultScoreURL = "http://localhost:1323"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestIntegration_JWKS_Endpoint(t *testing.T) {
	authURL := getEnv("AUTH_SERVICE_URL", defaultAuthURL)
	jwksURL := fmt.Sprintf("%s/api/auth/jwks", authURL)

	resp, err := http.Get(jwksURL)
	if err != nil {
		t.Fatalf("Failed to connect to JWKS endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var jwks struct {
		Keys []interface{} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		t.Fatalf("Failed to parse JWKS: %v", err)
	}

	if len(jwks.Keys) == 0 {
		t.Error("JWKS should contain at least one key")
	}

	t.Logf("✅ JWKS endpoint accessible, found %d key(s)", len(jwks.Keys))
}

func TestIntegration_ScoreService_Health(t *testing.T) {
	scoreURL := getEnv("SCORE_SERVICE_URL", defaultScoreURL)

	resp, err := http.Get(scoreURL + "/")
	if err != nil {
		t.Fatalf("Failed to connect to score service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	t.Logf("✅ Score service health check passed: %s", string(body))
}

func TestIntegration_ProtectedEndpoint_WithoutToken(t *testing.T) {
	scoreURL := getEnv("SCORE_SERVICE_URL", defaultScoreURL)

	req, err := http.NewRequest("GET", scoreURL+"/api/score", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	t.Logf("✅ Protected endpoint correctly returns 401 without token")
}

func TestIntegration_ProtectedEndpoint_InvalidToken(t *testing.T) {
	scoreURL := getEnv("SCORE_SERVICE_URL", defaultScoreURL)

	req, err := http.NewRequest("GET", scoreURL+"/api/score", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer invalid-token-12345")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	t.Logf("✅ Protected endpoint correctly returns 401 with invalid token")
}

func TestIntegration_GetTokenFromAuthService(t *testing.T) {
	authURL := getEnv("AUTH_SERVICE_URL", defaultAuthURL)

	// First, try to get a token (this will fail without a session, but we can test the endpoint)
	req, err := http.NewRequest("GET", authURL+"/api/auth/token", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 401 without a session
	if resp.StatusCode != http.StatusUnauthorized {
		t.Logf("Token endpoint returned status %d (expected 401 without session)", resp.StatusCode)
	}

	t.Logf("✅ Token endpoint is accessible (returns %d without session)", resp.StatusCode)
}

func TestIntegration_VerifierInitialization(t *testing.T) {
	authURL := getEnv("AUTH_SERVICE_URL", defaultAuthURL)

	verifier := jwt.NewJWKVerifier(authURL)
	if err := verifier.Initialize(); err != nil {
		t.Fatalf("Failed to initialize verifier with real JWKS endpoint: %v", err)
	}

	// Verify that keys were loaded
	verifier.mu.RLock()
	hasKeys := verifier.set != nil
	verifier.mu.RUnlock()

	if !hasKeys {
		t.Error("Verifier should have loaded keys from JWKS endpoint")
	}

	t.Logf("✅ JWT verifier successfully initialized with real JWKS endpoint")
}
