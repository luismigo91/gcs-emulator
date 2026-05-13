package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func newTestServer(t *testing.T) (*httptest.Server, backend.Backend) {
	t.Helper()
	b := backend.NewMemoryBackend()
	mux := router.New(b, nil, nil, nil, nil, nil, nil, "test-project")
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, b
}

func doRequest(t *testing.T, srv *httptest.Server, method, path string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("X-Goog-User-Project", "test-project")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	return resp
}

func TestBucketAPI(t *testing.T) {
	srv, _ := newTestServer(t)

	t.Run("POST /storage/v1/b — create bucket", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"test-bucket-1"}`)
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["name"] != "test-bucket-1" {
			t.Errorf("Expected name test-bucket-1, got %v", result["name"])
		}
	})

	t.Run("POST /storage/v1/b — duplicate bucket returns 409", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"test-bucket-1"}`)
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /storage/v1/b — list buckets", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"list-bucket-1"}`)
		doRequest(t, srv, http.MethodPost, "/storage/v1/b", body).Body.Close()
		body = bytes.NewBufferString(`{"name":"list-bucket-2"}`)
		doRequest(t, srv, http.MethodPost, "/storage/v1/b", body).Body.Close()

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b?project=test-project", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		items := result["items"].([]interface{})
		if len(items) < 2 {
			t.Errorf("Expected at least 2 buckets, got %d", len(items))
		}
	})

	t.Run("GET /storage/v1/b/{bucket} — get bucket metadata", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"get-bucket"}`)
		doRequest(t, srv, http.MethodPost, "/storage/v1/b", body).Body.Close()

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/get-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["name"] != "get-bucket" {
			t.Errorf("Expected name get-bucket, got %v", result["name"])
		}
	})

	t.Run("GET /storage/v1/b/{bucket} — missing bucket returns 404", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/nonexistent-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("DELETE /storage/v1/b/{bucket} — delete empty bucket", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"delete-empty-bucket"}`)
		doRequest(t, srv, http.MethodPost, "/storage/v1/b", body).Body.Close()

		resp := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/delete-empty-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("PATCH /storage/v1/b/{bucket} — update bucket metadata", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"update-bucket"}`)
		doRequest(t, srv, http.MethodPost, "/storage/v1/b", body).Body.Close()

		updateBody := bytes.NewBufferString(`{"labels":{"env":"test"}}`)
		resp := doRequest(t, srv, http.MethodPatch, "/storage/v1/b/update-bucket", updateBody)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestIAMPolicies(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "iam-bucket"}, backend.Conditions{})

	t.Run("GET /storage/v1/b/{bucket}/iam — get default policy", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/iam-bucket/iam", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["kind"] != "storage#policy" {
			t.Errorf("Expected kind storage#policy, got %v", result["kind"])
		}
		bindings, ok := result["bindings"].([]interface{})
		if !ok {
			t.Fatalf("Expected bindings array, got %v", result["bindings"])
		}
		if len(bindings) != 0 {
			t.Errorf("Expected empty bindings, got %d", len(bindings))
		}
	})

	t.Run("POST /storage/v1/b/{bucket}/iam — set policy", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"bindings": [
				{
					"role": "roles/storage.objectViewer",
					"members": ["user:test@example.com", "serviceAccount:svc@project.iam.gserviceaccount.com"]
				},
				{
					"role": "roles/storage.admin",
					"members": ["user:admin@example.com"]
				}
			]
		}`)
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/iam-bucket/iam", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		bindings, ok := result["bindings"].([]interface{})
		if !ok {
			t.Fatalf("Expected bindings array, got %v", result["bindings"])
		}
		if len(bindings) != 2 {
			t.Errorf("Expected 2 bindings, got %d", len(bindings))
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/iam — get updated policy", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/iam-bucket/iam", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		bindings, ok := result["bindings"].([]interface{})
		if !ok {
			t.Fatalf("Expected bindings array, got %v", result["bindings"])
		}
		if len(bindings) != 2 {
			t.Errorf("Expected 2 bindings after update, got %d", len(bindings))
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/iam:testIamPermissions — test permissions", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/iam-bucket/iam:testIamPermissions?permissions=storage.buckets.get&permissions=storage.objects.create&permissions=storage.objects.delete", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		permissions, ok := result["permissions"].([]interface{})
		if !ok {
			t.Fatalf("Expected permissions array, got %v", result["permissions"])
		}
		if len(permissions) != 3 {
			t.Errorf("Expected 3 permissions, got %d", len(permissions))
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/iam — nonexistent bucket returns 404", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/nonexistent-bucket/iam", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestNotifications(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "notif-bucket"}, backend.Conditions{})

	t.Run("POST /storage/v1/b/{bucket}/notificationConfigs — create notification", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"topic": "projects/test-project/topics/my-topic",
			"event_types": ["OBJECT_FINALIZE", "OBJECT_DELETE"],
			"payload_format": "JSON_API_V1"
		}`)
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/notif-bucket/notificationConfigs", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["kind"] != "storage#notification" {
			t.Errorf("Expected kind storage#notification, got %v", result["kind"])
		}
		if result["id"] == "" {
			t.Error("Expected notification ID to be generated")
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/notificationConfigs — list notifications", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/notif-bucket/notificationConfigs", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		items, ok := result["items"].([]interface{})
		if !ok {
			t.Fatalf("Expected items array, got %v", result["items"])
		}
		if len(items) < 1 {
			t.Errorf("Expected at least 1 notification, got %d", len(items))
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/notificationConfigs/{notification} — get notification", func(t *testing.T) {
		listResp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/notif-bucket/notificationConfigs", nil)
		var listResult map[string]interface{}
		json.NewDecoder(listResp.Body).Decode(&listResult)
		listResp.Body.Close()

		items := listResult["items"].([]interface{})
		notifID := items[0].(map[string]interface{})["id"].(string)

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/notif-bucket/notificationConfigs/"+notifID, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["id"] != notifID {
			t.Errorf("Expected id %s, got %v", notifID, result["id"])
		}
	})

	t.Run("DELETE /storage/v1/b/{bucket}/notificationConfigs/{notification} — delete notification", func(t *testing.T) {
		listResp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/notif-bucket/notificationConfigs", nil)
		var listResult map[string]interface{}
		json.NewDecoder(listResp.Body).Decode(&listResult)
		listResp.Body.Close()

		items := listResult["items"].([]interface{})
		notifID := items[0].(map[string]interface{})["id"].(string)

		resp := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/notif-bucket/notificationConfigs/"+notifID, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected 204, got %d", resp.StatusCode)
		}

		getResp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/notif-bucket/notificationConfigs/"+notifID, nil)
		defer getResp.Body.Close()
		if getResp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 after delete, got %d", getResp.StatusCode)
		}
	})

	t.Run("POST /storage/v1/b/{bucket}/notificationConfigs — nonexistent bucket returns 404", func(t *testing.T) {
		body := bytes.NewBufferString(`{"topic":"projects/test-project/topics/my-topic"}`)
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/nonexistent-bucket/notificationConfigs", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestSignedURLs(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "signed-bucket"}, backend.Conditions{})
	b.CreateObject(ctx, "signed-bucket", &model.Object{Name: "signed-obj.txt", Bucket: "signed-bucket", ContentType: "text/plain"}, strings.NewReader("signed content"), backend.Conditions{})

	t.Run("GET object with signed URL parameters — serves content", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/signed-bucket/o/signed-obj.txt?alt=media&X-Goog-Algorithm=GOOG4-RSA-SHA256&X-Goog-Credential=emulator-test%40project.iam.gserviceaccount.com%2F20260512%2Fauto%2Fstorage%2Fgoog4_request&X-Goog-Date=20260512T150405Z&X-Goog-Expires=3600&X-Goog-SignedHeaders=host&X-Goog-Signature=emulator-signature", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		data, _ := io.ReadAll(resp.Body)
		if string(data) != "signed content" {
			t.Errorf("Expected 'signed content', got '%s'", string(data))
		}
	})

	t.Run("GET object with custom signature — serves content", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/signed-bucket/o/signed-obj.txt?alt=media&X-Goog-Signature=custom-signature-123", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		data, _ := io.ReadAll(resp.Body)
		if string(data) != "signed content" {
			t.Errorf("Expected 'signed content', got '%s'", string(data))
		}
	})
}

func TestLifecycleRules(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "lifecycle-bucket"}, backend.Conditions{})

	t.Run("PATCH /storage/v1/b/{bucket} — set lifecycle rules", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"lifecycle": {
				"rule": [
					{
						"action": {"type": "Delete"},
						"condition": {"age": 30}
					},
					{
						"action": {"type": "SetStorageClass", "storageClass": "NEARLINE"},
						"condition": {"age": 7, "matchesStorageClass": ["STANDARD"]}
					}
				]
			}
		}`)
		resp := doRequest(t, srv, http.MethodPatch, "/storage/v1/b/lifecycle-bucket", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		lifecycle, ok := result["lifecycle"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected lifecycle object, got %v", result["lifecycle"])
		}
		rules, ok := lifecycle["rule"].([]interface{})
		if !ok {
			t.Fatalf("Expected rule array, got %v", lifecycle["rule"])
		}
		if len(rules) != 2 {
			t.Errorf("Expected 2 rules, got %d", len(rules))
		}
	})

	t.Run("GET /storage/v1/b/{bucket} — get lifecycle rules", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/lifecycle-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		lifecycle, ok := result["lifecycle"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected lifecycle object, got %v", result["lifecycle"])
		}
		rules, ok := lifecycle["rule"].([]interface{})
		if !ok {
			t.Fatalf("Expected rule array, got %v", lifecycle["rule"])
		}
		if len(rules) != 2 {
			t.Errorf("Expected 2 rules, got %d", len(rules))
		}
	})
}

func TestCORSConfiguration(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "cors-bucket"}, backend.Conditions{})

	t.Run("PATCH /storage/v1/b/{bucket} — set CORS rules", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"cors": [
				{
					"origin": ["https://example.com", "https://app.example.com"],
					"method": ["GET", "POST", "PUT"],
					"responseHeader": ["Content-Type", "X-Goog-Meta-*"],
					"maxAgeSeconds": 3600
				}
			]
		}`)
		resp := doRequest(t, srv, http.MethodPatch, "/storage/v1/b/cors-bucket", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		cors, ok := result["cors"].([]interface{})
		if !ok {
			t.Fatalf("Expected cors array, got %v", result["cors"])
		}
		if len(cors) != 1 {
			t.Errorf("Expected 1 CORS rule, got %d", len(cors))
		}
	})

	t.Run("GET /storage/v1/b/{bucket} — get CORS rules", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/cors-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		cors, ok := result["cors"].([]interface{})
		if !ok {
			t.Fatalf("Expected cors array, got %v", result["cors"])
		}
		if len(cors) != 1 {
			t.Errorf("Expected 1 CORS rule, got %d", len(cors))
		}
	})
}

func TestMetricsEndpoint(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "metrics-bucket"}, backend.Conditions{})
	b.CreateObject(ctx, "metrics-bucket", &model.Object{Name: "metrics-obj.txt", Bucket: "metrics-bucket"}, strings.NewReader("metrics content"), backend.Conditions{})

	t.Run("GET /metrics — returns Prometheus metrics", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/metrics", nil)
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "text/plain") {
			t.Errorf("Expected Content-Type text/plain, got %s", resp.Header.Get("Content-Type"))
		}

		data, _ := io.ReadAll(resp.Body)
		body := string(data)

		if !strings.Contains(body, "gcs_emulator_requests_total") {
			t.Error("Expected gcs_emulator_requests_total metric")
		}
		if !strings.Contains(body, "gcs_emulator_buckets_total") {
			t.Error("Expected gcs_emulator_buckets_total metric")
		}
		if !strings.Contains(body, "gcs_emulator_objects_total") {
			t.Error("Expected gcs_emulator_objects_total metric")
		}
	})
}

func TestBucketHEAD(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "head-bucket"}, backend.Conditions{})

	t.Run("HEAD /storage/v1/b/{bucket} — existing bucket returns 200", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodHead, srv.URL+"/storage/v1/b/head-bucket", nil)
		req.Header.Set("X-Goog-User-Project", "test-project")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(body))
		}

		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("HEAD /storage/v1/b/{bucket} — missing bucket returns 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodHead, srv.URL+"/storage/v1/b/nonexistent-bucket", nil)
		req.Header.Set("X-Goog-User-Project", "test-project")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(body))
		}
	})
}

func TestObjectAPI(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "test-bucket"}, backend.Conditions{})

	t.Run("POST /upload/storage/v1/b/{bucket}/o — create object", func(t *testing.T) {
		content := "hello world"
		resp := doRequest(t, srv, http.MethodPost, "/upload/storage/v1/b/test-bucket/o?name=test.txt", strings.NewReader(content))
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["name"] != "test.txt" {
			t.Errorf("Expected name test.txt, got %v", result["name"])
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/o — list objects", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "a.txt", Bucket: "test-bucket"}, strings.NewReader("a"), backend.Conditions{})
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "b.txt", Bucket: "test-bucket"}, strings.NewReader("b"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/test-bucket/o", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		items, ok := result["items"].([]interface{})
		if !ok || items == nil {
			t.Fatalf("Expected items array, got %v", result["items"])
		}
		if len(items) < 2 {
			t.Errorf("Expected at least 2 items, got %d", len(items))
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/o — list with prefix/delimiter", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "logs/2024/jan.txt", Bucket: "test-bucket"}, strings.NewReader("jan"), backend.Conditions{})
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "docs/readme.md", Bucket: "test-bucket"}, strings.NewReader("readme"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/test-bucket/o?prefix=logs/&delimiter=/", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		prefixes, ok := result["prefixes"].([]interface{})
		if !ok || prefixes == nil {
			t.Fatalf("Expected prefixes array, got %v", result["prefixes"])
		}
		if len(prefixes) != 1 {
			t.Errorf("Expected 1 prefix, got %d", len(prefixes))
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/o/{object} — get object metadata", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "meta.txt", Bucket: "test-bucket"}, strings.NewReader("data"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/test-bucket/o/meta.txt", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["name"] != "meta.txt" {
			t.Errorf("Expected name meta.txt, got %v", result["name"])
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/o/{object}?alt=media — download object", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "download.txt", Bucket: "test-bucket"}, strings.NewReader("download content"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/test-bucket/o/download.txt?alt=media", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		data, _ := io.ReadAll(resp.Body)
		if string(data) != "download content" {
			t.Errorf("Expected 'download content', got '%s'", string(data))
		}
	})

	t.Run("GET /storage/v1/b/{bucket}/o/{object}?alt=media — Range header", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "range.txt", Bucket: "test-bucket"}, strings.NewReader("0123456789"), backend.Conditions{})

		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/storage/v1/b/test-bucket/o/range.txt?alt=media", nil)
		req.Header.Set("X-Goog-User-Project", "test-project")
		req.Header.Set("Range", "bytes=2-5")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusPartialContent {
			t.Errorf("Expected 206, got %d", resp.StatusCode)
		}

		data, _ := io.ReadAll(resp.Body)
		if string(data) != "2345" {
			t.Errorf("Expected '2345', got '%s'", string(data))
		}

		if resp.Header.Get("Content-Range") == "" {
			t.Error("Expected Content-Range header")
		}
	})

	t.Run("DELETE /storage/v1/b/{bucket}/o/{object} — delete object", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "delete-me.txt", Bucket: "test-bucket"}, strings.NewReader("data"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/test-bucket/o/delete-me.txt", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("PATCH /storage/v1/b/{bucket}/o/{object} — update object metadata", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "update-obj.txt", Bucket: "test-bucket"}, strings.NewReader("data"), backend.Conditions{})

		updateBody := bytes.NewBufferString(`{"contentType":"text/plain"}`)
		resp := doRequest(t, srv, http.MethodPatch, "/storage/v1/b/test-bucket/o/update-obj.txt", updateBody)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("DELETE /storage/v1/b/{bucket} — non-empty bucket returns 409", func(t *testing.T) {
		b.CreateObject(ctx, "test-bucket", &model.Object{Name: "keep.txt", Bucket: "test-bucket"}, strings.NewReader("data"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/test-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestObjectHEAD(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "head-bucket"}, backend.Conditions{})
	b.CreateObject(ctx, "head-bucket", &model.Object{Name: "head-obj.txt", Bucket: "head-bucket", ContentType: "text/plain"}, strings.NewReader("hello world"), backend.Conditions{})

	t.Run("HEAD /storage/v1/b/{bucket}/o/{object} — existing object returns 200 with headers", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodHead, srv.URL+"/storage/v1/b/head-bucket/o/head-obj.txt", nil)
		req.Header.Set("X-Goog-User-Project", "test-project")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(body))
		}

		if resp.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", resp.Header.Get("Content-Type"))
		}

		if resp.Header.Get("Content-Length") != "11" {
			t.Errorf("Expected Content-Length 11, got %s", resp.Header.Get("Content-Length"))
		}

		if resp.Header.Get("X-Goog-Generation") == "" {
			t.Error("Expected X-Goog-Generation header")
		}

		if resp.Header.Get("X-Goog-Hash") == "" {
			t.Error("Expected X-Goog-Hash header")
		}

		if resp.Header.Get("ETag") == "" {
			t.Error("Expected ETag header")
		}

		if resp.Header.Get("Last-Modified") == "" {
			t.Error("Expected Last-Modified header")
		}
	})

	t.Run("HEAD /storage/v1/b/{bucket}/o/{object} — missing object returns 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodHead, srv.URL+"/storage/v1/b/head-bucket/o/nonexistent.txt", nil)
		req.Header.Set("X-Goog-User-Project", "test-project")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if len(body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(body))
		}
	})
}

func TestComposeCopyRewrite(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "ops-bucket"}, backend.Conditions{})
	b.CreateObject(ctx, "ops-bucket", &model.Object{Name: "src1.txt", Bucket: "ops-bucket"}, strings.NewReader("hello"), backend.Conditions{})
	b.CreateObject(ctx, "ops-bucket", &model.Object{Name: "src2.txt", Bucket: "ops-bucket"}, strings.NewReader(" world"), backend.Conditions{})

	t.Run("POST /storage/v1/b/{bucket}/o/{object}/compose — compose objects", func(t *testing.T) {
		body := bytes.NewBufferString(`{"sourceObjects":[{"name":"src1.txt"},{"name":"src2.txt"}]}`)
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/ops-bucket/o/combined.txt/compose", body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["componentCount"].(float64) != 2 {
			t.Errorf("Expected componentCount 2, got %v", result["componentCount"])
		}
	})

	t.Run("POST /storage/v1/b/{bucket}/o/{object}/copyTo/{destBucket}/o/{destObject} — copy object", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/ops-bucket/o/src1.txt/copyTo/ops-bucket/o/copied.txt", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["name"] != "copied.txt" {
			t.Errorf("Expected name copied.txt, got %v", result["name"])
		}
	})

	t.Run("POST /storage/v1/b/{bucket}/o/{object}/rewriteTo/{destBucket}/o/{destObject} — rewrite object", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/ops-bucket/o/src1.txt/rewriteTo/ops-bucket/o/rewritten.txt", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if result["done"] != true {
			t.Errorf("Expected done true, got %v", result["done"])
		}
	})
}

func TestUploads(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "upload-bucket"}, backend.Conditions{})

	t.Run("Simple upload via POST /upload/storage/v1/b/{bucket}/o", func(t *testing.T) {
		content := "simple upload content"
		resp := doRequest(t, srv, http.MethodPost, "/upload/storage/v1/b/upload-bucket/o?name=simple.txt", strings.NewReader(content))
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Multipart upload with metadata", func(t *testing.T) {
		var buf bytes.Buffer
		buf.WriteString("--boundary\r\n")
		buf.WriteString("Content-Type: application/json; charset=UTF-8\r\n\r\n")
		buf.WriteString(`{"name":"multipart.txt","metadata":{"key":"value"}}`)
		buf.WriteString("\r\n--boundary\r\n")
		buf.WriteString("Content-Type: text/plain\r\n\r\n")
		buf.WriteString("multipart content")
		buf.WriteString("\r\n--boundary--\r\n")

		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/upload/storage/v1/b/upload-bucket/o?name=multipart.txt", &buf)
		req.Header.Set("X-Goog-User-Project", "test-project")
		req.Header.Set("Content-Type", "multipart/related; boundary=boundary")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Resumable upload session creation", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/resumable/upload/storage/v1/b/upload-bucket/o?name=resumable.txt", nil)
		req.Header.Set("X-Goog-User-Project", "test-project")
		req.Header.Set("X-Goog-Upload-Command", "start")
		req.Header.Set("X-Goog-Upload-Protocol", "resumable")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		if resp.Header.Get("X-Goog-Upload-URL") == "" {
			t.Error("Expected X-Goog-Upload-URL header")
		}

		if resp.Header.Get("X-Goog-Upload-Status") != "active" {
			t.Errorf("Expected X-Goog-Upload-Status active, got %s", resp.Header.Get("X-Goog-Upload-Status"))
		}
	})

	t.Run("Resumable upload chunk upload and finalization", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/resumable/upload/storage/v1/b/upload-bucket/o?name=chunked.txt", nil)
		req.Header.Set("X-Goog-User-Project", "test-project")
		req.Header.Set("X-Goog-Upload-Command", "start")
		req.Header.Set("X-Goog-Upload-Protocol", "resumable")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for session creation, got %d", resp.StatusCode)
		}

		uploadPath := resp.Header.Get("X-Goog-Upload-URL")
		if uploadPath == "" {
			t.Fatal("Expected X-Goog-Upload-URL header")
		}

		uploadURL := srv.URL + uploadPath

		chunk1 := []byte("hello ")
		req, _ = http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(chunk1))
		req.Header.Set("X-Goog-User-Project", "test-project")
		req.Header.Set("Content-Range", fmt.Sprintf("bytes 0-%d/11", len(chunk1)-1))
		req.Header.Set("X-Goog-Upload-Command", "upload")
		resp, err = srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to upload chunk 1: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 for chunk 1, got %d", resp.StatusCode)
		}

		chunk2 := []byte("world")
		req, _ = http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(chunk2))
		req.Header.Set("X-Goog-User-Project", "test-project")
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-10/11", len(chunk1)))
		req.Header.Set("X-Goog-Upload-Command", "upload, finalize")
		resp, err = srv.Client().Do(req)
		if err != nil {
			t.Fatalf("Failed to finalize upload: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected 200 for finalize, got %d. Body: %s", resp.StatusCode, string(body))
		}

		if resp.Header.Get("X-Goog-Upload-Status") != "final" {
			t.Errorf("Expected X-Goog-Upload-Status final, got %s", resp.Header.Get("X-Goog-Upload-Status"))
		}

		content, err := b.GetObjectContent(ctx, "upload-bucket", "chunked.txt", 0)
		if err != nil {
			t.Fatalf("Failed to get object content: %v", err)
		}
		data, _ := io.ReadAll(content)
		if string(data) != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", string(data))
		}
	})
}

func TestXMLAPI(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "xml-bucket"}, backend.Conditions{})

	t.Run("PUT /{bucket}/{object} — upload object", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodPut, "/xml-bucket/xml-obj.txt", strings.NewReader("xml content"))
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("GET /{bucket}/{object} — download object", func(t *testing.T) {
		b.CreateObject(ctx, "xml-bucket", &model.Object{Name: "xml-get.txt", Bucket: "xml-bucket"}, strings.NewReader("get content"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/xml-bucket/xml-get.txt", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		data, _ := io.ReadAll(resp.Body)
		if string(data) != "get content" {
			t.Errorf("Expected 'get content', got '%s'", string(data))
		}
	})

	t.Run("GET /{bucket} — list objects (XML format)", func(t *testing.T) {
		b.CreateObject(ctx, "xml-bucket", &model.Object{Name: "xml-list.txt", Bucket: "xml-bucket"}, strings.NewReader("list content"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/xml-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/xml") {
			t.Errorf("Expected Content-Type application/xml, got %s", resp.Header.Get("Content-Type"))
		}

		data, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(data), "<ListBucketResult>") {
			t.Errorf("Expected XML ListBucketResult, got %s", string(data))
		}
	})

	t.Run("DELETE /{bucket}/{object} — delete object", func(t *testing.T) {
		b.CreateObject(ctx, "xml-bucket", &model.Object{Name: "xml-del.txt", Bucket: "xml-bucket"}, strings.NewReader("del content"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodDelete, "/xml-bucket/xml-del.txt", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("PUT /{bucket} — create bucket", func(t *testing.T) {
		resp := doRequest(t, srv, http.MethodPut, "/xml-new-bucket", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestVersioningAPI(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	t.Run("Overwrite preserves old generation", func(t *testing.T) {
		b.CreateBucket(ctx, &model.Bucket{
			Name:       "versioned-bucket-1",
			Versioning: &model.Versioning{Enabled: true},
		}, backend.Conditions{})

		b.CreateObject(ctx, "versioned-bucket-1", &model.Object{Name: "ver.txt", Bucket: "versioned-bucket-1"}, strings.NewReader("v1"), backend.Conditions{})
		b.CreateObject(ctx, "versioned-bucket-1", &model.Object{Name: "ver.txt", Bucket: "versioned-bucket-1"}, strings.NewReader("v2"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/versioned-bucket-1/o?versions=true", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		items, ok := result["items"].([]interface{})
		if !ok || items == nil {
			t.Fatalf("Expected items array, got %v", result["items"])
		}
		if len(items) != 2 {
			t.Errorf("Expected 2 versions, got %d", len(items))
		}
	})

	t.Run("Delete without generation preserves version", func(t *testing.T) {
		b.CreateBucket(ctx, &model.Bucket{
			Name:       "versioned-bucket-2",
			Versioning: &model.Versioning{Enabled: true},
		}, backend.Conditions{})

		b.CreateObject(ctx, "versioned-bucket-2", &model.Object{Name: "del-ver.txt", Bucket: "versioned-bucket-2"}, strings.NewReader("data"), backend.Conditions{})
		
		resp := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/versioned-bucket-2/o/del-ver.txt", nil)
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected delete to return 204, got %d", resp.StatusCode)
		}

		respAll, _ := b.ListObjects(ctx, "versioned-bucket-2", backend.ListObjectsParams{Versions: true})
		if len(respAll.Items) < 1 {
			t.Errorf("Expected at least 1 version after delete, got %d", len(respAll.Items))
		}

		respLive, _ := b.ListObjects(ctx, "versioned-bucket-2", backend.ListObjectsParams{})
		if len(respLive.Items) != 0 {
			t.Errorf("Expected 0 live versions after delete, got %d", len(respLive.Items))
		}
	})

	t.Run("Generation-specific get", func(t *testing.T) {
		b.CreateBucket(ctx, &model.Bucket{
			Name:       "versioned-bucket-3",
			Versioning: &model.Versioning{Enabled: true},
		}, backend.Conditions{})

		obj := &model.Object{Name: "gen.txt", Bucket: "versioned-bucket-3"}
		b.CreateObject(ctx, "versioned-bucket-3", obj, strings.NewReader("gen data"), backend.Conditions{})

		resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/versioned-bucket-3/o/gen.txt?generation="+strings.TrimSpace(""), nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := t.Context()

	b.CreateBucket(ctx, &model.Bucket{Name: "concurrent-bucket"}, backend.Conditions{})

	t.Run("Concurrent bucket creation — no duplicates", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(n int) {
				body := bytes.NewBufferString(`{"name":"unique-bucket"}`)
				resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b", body)
				resp.Body.Close()
				done <- resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict
			}(i)
		}

		for i := 0; i < 10; i++ {
			if !<-done {
				t.Error("Concurrent bucket creation returned unexpected status")
			}
		}
	})

	t.Run("Concurrent object writes — no data corruption", func(t *testing.T) {
		done := make(chan bool, 20)
		for i := 0; i < 20; i++ {
			go func(n int) {
				name := "obj-" + string(rune('a'+n)) + ".txt"
				content := "content-" + string(rune('a'+n))
				resp := doRequest(t, srv, http.MethodPost, "/upload/storage/v1/b/concurrent-bucket/o?name="+name, strings.NewReader(content))
				resp.Body.Close()
				done <- resp.StatusCode == http.StatusOK
			}(i)
		}

		for i := 0; i < 20; i++ {
			if !<-done {
				t.Error("Concurrent object write failed")
			}
		}
	})

	t.Run("Concurrent reads during writes", func(t *testing.T) {
		b.CreateObject(ctx, "concurrent-bucket", &model.Object{Name: "read-obj.txt", Bucket: "concurrent-bucket"}, strings.NewReader("initial"), backend.Conditions{})

		done := make(chan bool, 10)
		for i := 0; i < 5; i++ {
			go func() {
				resp := doRequest(t, srv, http.MethodGet, "/storage/v1/b/concurrent-bucket/o/read-obj.txt?alt=media", nil)
				resp.Body.Close()
				done <- resp.StatusCode == http.StatusOK
			}()
		}
		for i := 0; i < 5; i++ {
			go func() {
				resp := doRequest(t, srv, http.MethodPost, "/upload/storage/v1/b/concurrent-bucket/o?name=read-obj.txt", strings.NewReader("updated"))
				resp.Body.Close()
				done <- resp.StatusCode == http.StatusOK
			}()
		}

		for i := 0; i < 10; i++ {
			if !<-done {
				t.Error("Concurrent read/write failed")
			}
		}
	})
}

func TestDataPreloading(t *testing.T) {
	seedDir := t.TempDir()
	bucketDir := seedDir + "/seed-bucket"
	if err := os.MkdirAll(bucketDir+"/subdir", 0755); err != nil {
		t.Fatalf("Failed to create seed directory: %v", err)
	}
	if err := os.WriteFile(bucketDir+"/file.txt", []byte("seeded content"), 0644); err != nil {
		t.Fatalf("Failed to create seed file: %v", err)
	}
	if err := os.WriteFile(bucketDir+"/subdir/nested.json", []byte(`{"key":"value"}`), 0644); err != nil {
		t.Fatalf("Failed to create nested seed file: %v", err)
	}

	b := backend.NewMemoryBackend()
	if err := backend.SeedFromDirectory(b, seedDir); err != nil {
		t.Fatalf("SeedFromDirectory failed: %v", err)
	}

	ctx := t.Context()

	t.Run("Bucket created from seed directory", func(t *testing.T) {
		_, err := b.GetBucket(ctx, "seed-bucket")
		if err != nil {
			t.Errorf("Expected seed bucket to exist, got error: %v", err)
		}
	})

	t.Run("Objects created from seed files", func(t *testing.T) {
		obj, err := b.GetObject(ctx, "seed-bucket", "file.txt", 0)
		if err != nil {
			t.Fatalf("Expected file.txt to exist, got error: %v", err)
		}
		if obj.ContentType != "text/plain" {
			t.Errorf("Expected content-type text/plain, got %s", obj.ContentType)
		}

		obj, err = b.GetObject(ctx, "seed-bucket", "subdir/nested.json", 0)
		if err != nil {
			t.Fatalf("Expected subdir/nested.json to exist, got error: %v", err)
		}
		if obj.ContentType != "application/json" {
			t.Errorf("Expected content-type application/json, got %s", obj.ContentType)
		}
	})

	t.Run("Object content matches seed file", func(t *testing.T) {
		content, err := b.GetObjectContent(ctx, "seed-bucket", "file.txt", 0)
		if err != nil {
			t.Fatalf("Failed to get object content: %v", err)
		}
		data, _ := io.ReadAll(content)
		if string(data) != "seeded content" {
			t.Errorf("Expected 'seeded content', got '%s'", string(data))
		}
	})
}

func TestGracefulShutdown(t *testing.T) {
	t.Run("Persistent backend saves on shutdown", func(t *testing.T) {
		storagePath := t.TempDir()
		b, err := backend.NewBackend(backend.BackendConfig{
			Mode:        backend.ModePersistent,
			StoragePath: storagePath,
		})
		if err != nil {
			t.Fatalf("Failed to create persistent backend: %v", err)
		}

		ctx := context.Background()
		b.CreateBucket(ctx, &model.Bucket{Name: "persistent-bucket"}, backend.Conditions{})
		b.CreateObject(ctx, "persistent-bucket", &model.Object{Name: "test.txt", Bucket: "persistent-bucket"}, strings.NewReader("persistent data"), backend.Conditions{})

		if err := b.Shutdown(); err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}

		dataFile := storagePath + "/data.json"
		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			t.Error("Expected data.json to exist after shutdown")
		}

		b2, err := backend.NewBackend(backend.BackendConfig{
			Mode:        backend.ModePersistent,
			StoragePath: storagePath,
		})
		if err != nil {
			t.Fatalf("Failed to reload persistent backend: %v", err)
		}

		obj, err := b2.GetObject(ctx, "persistent-bucket", "test.txt", 0)
		if err != nil {
			t.Errorf("Expected test.txt to exist after reload, got: %v", err)
		}
		if obj.Name != "test.txt" {
			t.Errorf("Expected name test.txt, got %s", obj.Name)
		}
	})

	t.Run("Hybrid backend flushes on shutdown", func(t *testing.T) {
		storagePath := t.TempDir()
		b, err := backend.NewBackend(backend.BackendConfig{
			Mode:          backend.ModeHybrid,
			StoragePath:   storagePath,
			FlushInterval: time.Hour,
		})
		if err != nil {
			t.Fatalf("Failed to create hybrid backend: %v", err)
		}

		ctx := context.Background()
		b.CreateBucket(ctx, &model.Bucket{Name: "hybrid-bucket"}, backend.Conditions{})
		b.CreateObject(ctx, "hybrid-bucket", &model.Object{Name: "test.txt", Bucket: "hybrid-bucket"}, strings.NewReader("hybrid data"), backend.Conditions{})

		if err := b.Shutdown(); err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}

		dataFile := storagePath + "/data.json"
		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			t.Error("Expected data.json to exist after shutdown")
		}

		b2, err := backend.NewBackend(backend.BackendConfig{
			Mode:          backend.ModeHybrid,
			StoragePath:   storagePath,
			FlushInterval: time.Hour,
		})
		if err != nil {
			t.Fatalf("Failed to reload hybrid backend: %v", err)
		}

		obj, err := b2.GetObject(ctx, "hybrid-bucket", "test.txt", 0)
		if err != nil {
			t.Errorf("Expected test.txt to exist after reload, got: %v", err)
		}
		if obj.Name != "test.txt" {
			t.Errorf("Expected name test.txt, got %s", obj.Name)
		}
	})

	t.Run("WAL backend compacts on shutdown", func(t *testing.T) {
		storagePath := t.TempDir()
		b, err := backend.NewBackend(backend.BackendConfig{
			Mode:        backend.ModeWAL,
			StoragePath: storagePath,
		})
		if err != nil {
			t.Fatalf("Failed to create WAL backend: %v", err)
		}

		ctx := context.Background()
		b.CreateBucket(ctx, &model.Bucket{Name: "wal-bucket"}, backend.Conditions{})
		b.CreateObject(ctx, "wal-bucket", &model.Object{Name: "test.txt", Bucket: "wal-bucket"}, strings.NewReader("wal data"), backend.Conditions{})

		walFile := storagePath + "/operations.wal"
		if _, err := os.Stat(walFile); os.IsNotExist(err) {
			t.Error("Expected operations.wal to exist before shutdown")
		}

		if err := b.Shutdown(); err != nil {
			t.Fatalf("Shutdown failed: %v", err)
		}

		snapshotFile := storagePath + "/snapshot.json"
		if _, err := os.Stat(snapshotFile); os.IsNotExist(err) {
			t.Error("Expected snapshot.json to exist after shutdown compaction")
		}

		_, err = backend.NewBackend(backend.BackendConfig{
			Mode:        backend.ModeWAL,
			StoragePath: storagePath,
		})
		if err != nil {
			t.Fatalf("Failed to reload WAL backend: %v", err)
		}
	})
}

func TestHealthCheck(t *testing.T) {
	srv, _ := newTestServer(t)

	resp := doRequest(t, srv, http.MethodGet, "/-/health", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", result["status"])
	}

	services, ok := result["services"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected services object in health response")
	}
	if services["gcs"] != "available" {
		t.Errorf("Expected gcs 'available', got '%v'", services["gcs"])
	}
}

func TestLifecycleAPI(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "lifecycle-bucket"}, backend.Conditions{})

	body := strings.NewReader(`{"rule":[{"action":{"type":"Delete"},"condition":{"age":30}}]}`)
	resp := doRequest(t, srv, http.MethodPatch, "/storage/v1/b/lifecycle-bucket/lifecycle", body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 setting lifecycle, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/storage/v1/b/lifecycle-bucket/lifecycle", nil)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting lifecycle, got %d", resp2.StatusCode)
	}

	var lifecycle model.Lifecycle
	json.NewDecoder(resp2.Body).Decode(&lifecycle)

	if len(lifecycle.Rule) != 1 {
		t.Errorf("Expected 1 lifecycle rule, got %d", len(lifecycle.Rule))
	}
	if lifecycle.Rule[0].Action.Type != "Delete" {
		t.Errorf("Expected action type 'Delete', got '%s'", lifecycle.Rule[0].Action.Type)
	}

	resp3 := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/lifecycle-bucket/lifecycle", nil)
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusNoContent {
		t.Errorf("Expected 204 deleting lifecycle, got %d", resp3.StatusCode)
	}

	resp4 := doRequest(t, srv, http.MethodGet, "/storage/v1/b/nonexistent/lifecycle", nil)
	defer resp4.Body.Close()

	if resp4.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for nonexistent bucket, got %d", resp4.StatusCode)
	}
}

func TestCORSAPI(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "cors-bucket"}, backend.Conditions{})

	body := strings.NewReader(`[{"origin":["*"],"method":["GET"],"maxAgeSeconds":3600}]`)
	resp := doRequest(t, srv, http.MethodPatch, "/storage/v1/b/cors-bucket/cors", body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 setting CORS, got %d: %s", resp.StatusCode, resp.Body)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/storage/v1/b/cors-bucket/cors", nil)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting CORS, got %d", resp2.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&result)

	items, ok := result["items"].([]interface{})
	if !ok || len(items) != 1 {
		t.Errorf("Expected 1 CORS rule, got %v", result["items"])
	}

	resp3 := doRequest(t, srv, http.MethodGet, "/storage/v1/b/nonexistent/cors", nil)
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for nonexistent bucket, got %d", resp3.StatusCode)
	}
}

func TestXMLDeleteBucket(t *testing.T) {
	srv, b := newTestServer(t)
	ctx := context.Background()

	b.CreateBucket(ctx, &model.Bucket{Name: "xml-del-bucket"}, backend.Conditions{})

	resp := doRequest(t, srv, http.MethodDelete, "/xml-del-bucket", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected 204 deleting empty bucket via XML, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/xml-del-bucket", nil)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for deleted bucket, got %d", resp2.StatusCode)
	}

	b.CreateBucket(ctx, &model.Bucket{Name: "xml-nonempty"}, backend.Conditions{})
	b.CreateObject(ctx, "xml-nonempty", &model.Object{Name: "obj.txt", Bucket: "xml-nonempty"}, strings.NewReader("data"), backend.Conditions{})

	resp3 := doRequest(t, srv, http.MethodDelete, "/xml-nonempty", nil)
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 deleting non-empty bucket via XML, got %d", resp3.StatusCode)
	}
}
