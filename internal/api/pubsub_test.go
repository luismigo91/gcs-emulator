package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	pubsubmodel "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func newPubSubTestServer(t *testing.T) (*httptest.Server, pubsub.PubSubBackend) {
	t.Helper()
	b := backend.NewMemoryBackend()
	psb := pubsub.NewMemoryPubSubBackend()
	mux := router.New(b, psb, nil, nil, nil, "test-project")
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, psb
}

func pubsubPath(path string) string {
	return path
}

func TestPubSubTopicCRUD(t *testing.T) {
	srv, _ := newPubSubTestServer(t)

	resp := doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/topics/test-topic"), strings.NewReader(`{"name":"projects/test-project/topics/test-topic"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating topic, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, srv, http.MethodGet, pubsubPath("/v1/projects/test-project/topics/test-topic"), nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 getting topic, got %d", resp2.StatusCode)
	}

	resp3 := doRequest(t, srv, http.MethodGet, pubsubPath("/v1/projects/test-project/topics"), nil)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 listing topics, got %d", resp3.StatusCode)
	}
}

func TestPubSubPublishPullAck(t *testing.T) {
	srv, _ := newPubSubTestServer(t)

	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/topics/flow-topic"), strings.NewReader(`{"name":"projects/test-project/topics/flow-topic"}`)).Body.Close()

	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/subscriptions/flow-sub"), strings.NewReader(`{"name":"projects/test-project/subscriptions/flow-sub","topic":"projects/test-project/topics/flow-topic"}`)).Body.Close()

	resp := doRequest(t, srv, http.MethodPost, pubsubPath("/v1/projects/test-project/topics/flow-topic:publish"), strings.NewReader(`{"messages":[{"data":"aGVsbG8="}]}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 publishing, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, srv, http.MethodPost, pubsubPath("/v1/projects/test-project/subscriptions/flow-sub:pull"), strings.NewReader(`{"maxMessages":10}`))
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 pulling, got %d", resp2.StatusCode)
	}

	var pullResp pubsubmodel.PullResponse
	json.NewDecoder(resp2.Body).Decode(&pullResp)
	if len(pullResp.ReceivedMessages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(pullResp.ReceivedMessages))
	}
	ackID := pullResp.ReceivedMessages[0].AckID
	if ackID == "" {
		t.Fatal("Expected non-empty ack ID")
	}

	resp3 := doRequest(t, srv, http.MethodPost, pubsubPath("/v1/projects/test-project/subscriptions/flow-sub:acknowledge"), strings.NewReader(`{"ackIds":["`+ackID+`"]}`))
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 acking, got %d", resp3.StatusCode)
	}
}

func TestPubSubSchemaCRUD(t *testing.T) {
	srv, _ := newPubSubTestServer(t)

	body := `{"name":"projects/test-project/schemas/test-schema","type":"AVRO","definition":"{\"type\":\"record\",\"name\":\"T\",\"fields\":[]}"}`
	resp := doRequest(t, srv, http.MethodPost, pubsubPath("/v1/projects/test-project/schemas"), strings.NewReader(body))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating schema, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, srv, http.MethodGet, pubsubPath("/v1/projects/test-project/schemas/test-schema"), nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting schema, got %d", resp2.StatusCode)
	}

	resp3 := doRequest(t, srv, http.MethodDelete, pubsubPath("/v1/projects/test-project/schemas/test-schema"), nil)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 deleting schema, got %d", resp3.StatusCode)
	}
}

func TestGCSNotificationTriggersPubSub(t *testing.T) {
	srv, psb := newPubSubTestServer(t)

	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/topics/gcs-events"), strings.NewReader(`{"name":"projects/test-project/topics/gcs-events"}`)).Body.Close()
	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/subscriptions/gcs-sub"), strings.NewReader(`{"name":"projects/test-project/subscriptions/gcs-sub","topic":"projects/test-project/topics/gcs-events"}`)).Body.Close()

	doRequest(t, srv, http.MethodPost, pubsubPath("/storage/v1/b"), strings.NewReader(`{"name":"notif-bucket"}`)).Body.Close()

	notifBody := `{"topic":"projects/test-project/topics/gcs-events","event_types":["OBJECT_FINALIZE"]}`
	doRequest(t, srv, http.MethodPost, pubsubPath("/storage/v1/b/notif-bucket/notificationConfigs"), strings.NewReader(notifBody)).Body.Close()

	resp := doRequest(t, srv, http.MethodPost, pubsubPath("/upload/storage/v1/b/notif-bucket/o?name=test.txt"), strings.NewReader("hello world"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating object, got %d: %s", resp.StatusCode, resp.Body)
	}

	msgs, err := psb.Pull(nil, "test-project", "gcs-sub", 10)
	if err != nil {
		t.Fatalf("Failed to pull: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 notification message, got %d", len(msgs))
	}
	if msgs[0].Message.Attributes["eventType"] != "OBJECT_FINALIZE" {
		t.Errorf("Expected eventType OBJECT_FINALIZE, got %s", msgs[0].Message.Attributes["eventType"])
	}
	if msgs[0].Message.Attributes["bucketId"] != "notif-bucket" {
		t.Errorf("Expected bucket notif-bucket, got %s", msgs[0].Message.Attributes["bucketId"])
	}
}

func TestBucketACL(t *testing.T) {
	srv, _ := newPubSubTestServer(t)

	doRequest(t, srv, http.MethodPost, "/storage/v1/b", strings.NewReader(`{"name":"acl-bucket"}`)).Body.Close()

	aclBody := `{"entity":"user-test@example.com","role":"READER"}`
	resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/acl-bucket/acl", strings.NewReader(aclBody))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating ACL, got %d: %s", resp.StatusCode, resp.Body)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/storage/v1/b/acl-bucket/acl", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 listing ACLs, got %d", resp2.StatusCode)
	}

	resp3 := doRequest(t, srv, http.MethodGet, "/storage/v1/b/acl-bucket/acl/user-test@example.com", nil)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 getting ACL, got %d", resp3.StatusCode)
	}

	resp4 := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/acl-bucket/acl/user-test@example.com", nil)
	defer resp4.Body.Close()
	if resp4.StatusCode != http.StatusNoContent {
		t.Errorf("Expected 204 deleting ACL, got %d", resp4.StatusCode)
	}
}

func TestObjectACL(t *testing.T) {
	srv, _ := newPubSubTestServer(t)

	doRequest(t, srv, http.MethodPost, "/storage/v1/b", strings.NewReader(`{"name":"oacl-bucket"}`)).Body.Close()
	doRequest(t, srv, http.MethodPost, "/upload/storage/v1/b/oacl-bucket/o?name=test.txt", strings.NewReader("data")).Body.Close()

	aclBody := `{"entity":"user-owner@example.com","role":"OWNER"}`
	resp := doRequest(t, srv, http.MethodPost, "/storage/v1/b/oacl-bucket/o/test.txt/acl", strings.NewReader(aclBody))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 creating object ACL, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, srv, http.MethodGet, "/storage/v1/b/oacl-bucket/o/test.txt/acl", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 listing object ACLs, got %d", resp2.StatusCode)
	}

	resp3 := doRequest(t, srv, http.MethodDelete, "/storage/v1/b/oacl-bucket/o/test.txt/acl/user-owner@example.com", nil)
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusNoContent {
		t.Errorf("Expected 204 deleting object ACL, got %d", resp3.StatusCode)
	}
}

func TestPubSubDeadLetterPolicy(t *testing.T) {
	srv, psb := newPubSubTestServer(t)

	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/topics/d-main"), strings.NewReader(`{"name":"projects/test-project/topics/d-main"}`)).Body.Close()
	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/topics/d-dlq"), strings.NewReader(`{"name":"projects/test-project/topics/d-dlq"}`)).Body.Close()
	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/subscriptions/d-dlq-sub"), strings.NewReader(`{"name":"projects/test-project/subscriptions/d-dlq-sub","topic":"projects/test-project/topics/d-dlq"}`)).Body.Close()

	subBody := `{"name":"projects/test-project/subscriptions/d-main-sub","topic":"projects/test-project/topics/d-main","ackDeadlineSeconds":1,"deadLetterPolicy":{"deadLetterTopic":"projects/test-project/topics/d-dlq","maxDeliveryAttempts":2}}`
	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/subscriptions/d-main-sub"), strings.NewReader(subBody)).Body.Close()

	psb.Publish(nil, "test-project", "d-main", []*pubsubmodel.PubSubMessage{{Data: []byte("dlq-test")}})

	time.Sleep(1100 * time.Millisecond)
	psb.Pull(nil, "test-project", "d-main-sub", 10)
	time.Sleep(1100 * time.Millisecond)
	psb.Pull(nil, "test-project", "d-main-sub", 10)

	r, err := psb.Pull(nil, "test-project", "d-dlq-sub", 10)
	if err != nil {
		t.Fatalf("Failed to pull from DLQ: %v", err)
	}
	if len(r) != 1 {
		t.Fatalf("Expected 1 DLQ message, got %d", len(r))
	}
	if string(r[0].Message.Data) != "dlq-test" {
		t.Errorf("Expected 'dlq-test', got '%s'", string(r[0].Message.Data))
	}
}

func TestPubSubMessageFilter(t *testing.T) {
	srv, psb := newPubSubTestServer(t)

	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/topics/filter-topic"), strings.NewReader(`{"name":"projects/test-project/topics/filter-topic"}`)).Body.Close()

	subBody := `{"name":"projects/test-project/subscriptions/filter-sub","topic":"projects/test-project/topics/filter-topic","filter":"attributes.env = \"prod\""}`
	doRequest(t, srv, http.MethodPut, pubsubPath("/v1/projects/test-project/subscriptions/filter-sub"), strings.NewReader(subBody)).Body.Close()

	doRequest(t, srv, http.MethodPost, pubsubPath("/v1/projects/test-project/topics/filter-topic:publish"), strings.NewReader(`{"messages":[{"data":"cHJvZA==","attributes":{"env":"prod"}},{"data":"ZGV2","attributes":{"env":"dev"}}]}`)).Body.Close()

	msgs, err := psb.Pull(nil, "test-project", "filter-sub", 10)
	if err != nil {
		t.Fatalf("Failed to pull: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 filtered message, got %d", len(msgs))
	}
	if string(msgs[0].Message.Data) != "prod" {
		t.Errorf("Expected 'prod', got '%s'", string(msgs[0].Message.Data))
	}
}
