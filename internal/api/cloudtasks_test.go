package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	tasksbackend "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/backend"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	secretmanager "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func newCloudTasksTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	b := backend.NewMemoryBackend()
	psb := pubsub.NewMemoryPubSubBackend()
	smb := secretmanager.NewMemorySecretManagerBackend()
	ctb := tasksbackend.NewMemoryCloudTasksBackend()
	mux := router.New(b, psb, smb, ctb, nil, nil, nil, "test-project")
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestCloudTasksQueueCRUD(t *testing.T) {
	srv := newCloudTasksTestServer(t)

	resp := doRequest(t, srv, http.MethodPost, "/v2/projects/test-project/locations/us-central1/queues?queueId=my-queue", strings.NewReader("{}"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating queue, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/v2/projects/test-project/locations/us-central1/queues/my-queue", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting queue, got %d", resp2.StatusCode)
	}

	resp3 := doRequest(t, srv, http.MethodGet, "/v2/projects/test-project/locations/us-central1/queues", nil)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 listing queues, got %d", resp3.StatusCode)
	}

	resp4 := doRequest(t, srv, http.MethodDelete, "/v2/projects/test-project/locations/us-central1/queues/my-queue", nil)
	defer resp4.Body.Close()
	if resp4.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 deleting queue, got %d", resp4.StatusCode)
	}
}

func TestCloudTasksTaskCRUD(t *testing.T) {
	srv := newCloudTasksTestServer(t)

	doRequest(t, srv, http.MethodPost, "/v2/projects/test-project/locations/us-central1/queues?queueId=tq", strings.NewReader("{}")).Body.Close()

	body := `{"task":{"httpRequest":{"url":"http://localhost:9999/test","httpMethod":"POST","body":"dGVzdA=="}}}`
	resp := doRequest(t, srv, http.MethodPost, "/v2/projects/test-project/locations/us-central1/queues/tq/tasks", strings.NewReader(body))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating task, got %d: %s", resp.StatusCode, resp.Body)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/v2/projects/test-project/locations/us-central1/queues/tq/tasks", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 listing tasks, got %d", resp2.StatusCode)
	}
}

func TestServicesEndpoint(t *testing.T) {
	srv := newCloudTasksTestServer(t)

	resp := doRequest(t, srv, http.MethodGet, "/__/services", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
}
