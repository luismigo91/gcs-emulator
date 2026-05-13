package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	secretmanager "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func newSecretManagerTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	b := backend.NewMemoryBackend()
	psb := pubsub.NewMemoryPubSubBackend()
	smb := secretmanager.NewMemorySecretManagerBackend()
	mux := router.New(b, psb, smb, nil, nil, nil, "test-project")
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestSecretManagerCRUD(t *testing.T) {
	srv := newSecretManagerTestServer(t)

	resp := doRequest(t, srv, http.MethodPost, "/v1/projects/test-project/secrets?secretId=my-secret", strings.NewReader(`{"replication":{"automatic":{}}}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating secret, got %d: %s", resp.StatusCode, resp.Body)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/v1/projects/test-project/secrets/my-secret", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting secret, got %d", resp2.StatusCode)
	}

	resp3 := doRequest(t, srv, http.MethodGet, "/v1/projects/test-project/secrets", nil)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 listing secrets, got %d", resp3.StatusCode)
	}

	resp4 := doRequest(t, srv, http.MethodDelete, "/v1/projects/test-project/secrets/my-secret", nil)
	defer resp4.Body.Close()
	if resp4.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 deleting secret, got %d", resp4.StatusCode)
	}
}

func TestSecretManagerVersions(t *testing.T) {
	srv := newSecretManagerTestServer(t)

	doRequest(t, srv, http.MethodPost, "/v1/projects/test-project/secrets?secretId=ver-secret", strings.NewReader(`{"replication":{"automatic":{}}}`)).Body.Close()

	resp := doRequest(t, srv, http.MethodPost, "/v1/projects/test-project/secrets/ver-secret:addVersion", strings.NewReader(`{"payload":{"data":"c3VwZXIgc2VjcmV0"}}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 adding version, got %d: %s", resp.StatusCode, resp.Body)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/v1/projects/test-project/secrets/ver-secret/versions", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 listing versions, got %d", resp2.StatusCode)
	}

	resp3 := doRequest(t, srv, http.MethodGet, "/v1/projects/test-project/secrets/ver-secret/versions/1:access", nil)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 accessing version, got %d: %s", resp3.StatusCode, resp3.Body)
	}

	resp4 := doRequest(t, srv, http.MethodPost, "/v1/projects/test-project/secrets/ver-secret/versions/1:disable", nil)
	defer resp4.Body.Close()
	if resp4.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 disabling version, got %d", resp4.StatusCode)
	}

	resp5 := doRequest(t, srv, http.MethodPost, "/v1/projects/test-project/secrets/ver-secret/versions/1:enable", nil)
	defer resp5.Body.Close()
	if resp5.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 enabling version, got %d", resp5.StatusCode)
	}
}
