package auth

import (
	"encoding/json"
	"net/http"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: "ya29.emulator.mock.access.token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "https://www.googleapis.com/auth/devstorage.full_control https://www.googleapis.com/auth/cloud-platform",
	})
}
