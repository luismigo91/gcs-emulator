package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/auth"
)

func TestTokenEndpoint_FormEncoded(t *testing.T) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", "fake-jwt-assertion")

	req := httptest.NewRequest(http.MethodPost, "/oauth2/v4/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	auth.TokenHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp auth.TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("access_token should not be empty")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("Expected token_type 'Bearer', got '%s'", resp.TokenType)
	}
	if resp.ExpiresIn <= 0 {
		t.Errorf("Expected positive expires_in, got %d", resp.ExpiresIn)
	}
}

func TestTokenEndpoint_JSONBody(t *testing.T) {
	body := `{"grant_type":"authorization_code","code":"fake-code"}`
	req := httptest.NewRequest(http.MethodPost, "/oauth2/v4/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	auth.TokenHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp auth.TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("access_token should not be empty")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("Expected token_type 'Bearer', got '%s'", resp.TokenType)
	}
}

func TestTokenEndpoint_GETReturnsMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/oauth2/v4/token", nil)
	w := httptest.NewRecorder()

	auth.TokenHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}
}

func TestTokenEndpoint_NoContentType(t *testing.T) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", "fake-refresh")

	req := httptest.NewRequest(http.MethodPost, "/oauth2/v4/token", strings.NewReader(form.Encode()))
	w := httptest.NewRecorder()

	auth.TokenHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp auth.TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("access_token should not be empty")
	}
}
