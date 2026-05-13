package api_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func TestHealthReportsAllServices(t *testing.T) {
	cfg := router.RouterConfig{DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	resp := doRequest(t, th.Server, http.MethodGet, "/-/health", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["status"] != "healthy" {
		t.Errorf("Expected status healthy")
	}
	services := result["services"].(map[string]interface{})
	if len(services) == 0 {
		t.Error("Expected services in health response")
	}
}

func TestDashboardReturnsHTML(t *testing.T) {
	cfg := router.RouterConfig{DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	resp := doRequest(t, th.Server, http.MethodGet, "/-/", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Expected text/html, got %s", ct)
	}
}

func TestServicesEndpointV2(t *testing.T) {
	cfg := router.RouterConfig{DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	resp := doRequest(t, th.Server, http.MethodGet, "/__/services", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["emulator"] != "gcs-emulator" {
		t.Errorf("Expected gcs-emulator, got %v", result["emulator"])
	}
}

func TestMetricsEndpointV2(t *testing.T) {
	cfg := router.RouterConfig{DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	resp := doRequest(t, th.Server, http.MethodGet, "/metrics", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
}
