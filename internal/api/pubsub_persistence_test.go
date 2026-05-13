package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func TestPubSubPersistent_SurvivesRestart(t *testing.T) {
	dir := t.TempDir()

	psb1, err := pubsub.NewPubSubBackend("persistent", dir)
	if err != nil {
		t.Fatalf("Failed to create persistent: %v", err)
	}
	mux1 := router.New(nil, psb1, nil, nil, nil, nil, "test-project")
	srv1 := httptest.NewServer(mux1)

	doRequest(t, srv1, http.MethodPut, "/v1/projects/test-project/topics/persist-topic", strings.NewReader(`{"name":"projects/test-project/topics/persist-topic"}`)).Body.Close()
	doRequest(t, srv1, http.MethodPut, "/v1/projects/test-project/subscriptions/persist-sub", strings.NewReader(`{"name":"projects/test-project/subscriptions/persist-sub","topic":"projects/test-project/topics/persist-topic"}`)).Body.Close()
	psb1.Shutdown()
	srv1.Close()

	psb2, err := pubsub.NewPubSubBackend("persistent", dir)
	if err != nil {
		t.Fatalf("Failed to recreate persistent: %v", err)
	}
	defer psb2.Shutdown()
	mux2 := router.New(nil, psb2, nil, nil, nil, nil, "test-project")
	srv2 := httptest.NewServer(mux2)
	defer srv2.Close()

	resp := doRequest(t, srv2, http.MethodGet, "/v1/projects/test-project/topics/persist-topic", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting topic after restart, got %d", resp.StatusCode)
	}
}

func TestPubSubWAL_SurvivesRestart(t *testing.T) {
	dir := t.TempDir()

	psb1, err := pubsub.NewPubSubBackend("wal", dir)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	mux1 := router.New(nil, psb1, nil, nil, nil, nil, "test-project")
	srv1 := httptest.NewServer(mux1)

	doRequest(t, srv1, http.MethodPut, "/v1/projects/test-project/topics/wal-topic", strings.NewReader(`{"name":"projects/test-project/topics/wal-topic"}`)).Body.Close()
	psb1.Shutdown()
	srv1.Close()

	psb2, err := pubsub.NewPubSubBackend("wal", dir)
	if err != nil {
		t.Fatalf("Failed to recreate WAL: %v", err)
	}
	defer psb2.Shutdown()
	mux2 := router.New(nil, psb2, nil, nil, nil, nil, "test-project")
	srv2 := httptest.NewServer(mux2)
	defer srv2.Close()

	resp := doRequest(t, srv2, http.MethodGet, "/v1/projects/test-project/topics/wal-topic", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting WAL topic after restart, got %d", resp.StatusCode)
	}
}
