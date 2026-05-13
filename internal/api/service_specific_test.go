package api_test

import (
	"net/http"
	"strings"
	"testing"
	"time"
	tasks "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/backend"
	kms "github.com/luismiguelgilolivert/gcs-emulator/internal/kms/backend"
	logging "github.com/luismiguelgilolivert/gcs-emulator/internal/logging/backend"
	monitoring "github.com/luismiguelgilolivert/gcs-emulator/internal/monitoring/backend"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	pubsubmodel "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

func TestBigQuerySQL(t *testing.T) {
	cfg := router.RouterConfig{DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	dsBody := `{"datasetReference":{"datasetId":"testds","projectId":"test-project"}}`
	doRequest(t, th.Server, http.MethodPost, "/bigquery/v2/projects/test-project/datasets", strings.NewReader(dsBody)).Body.Close()

	tblBody := `{"tableReference":{"datasetId":"testds","projectId":"test-project","tableId":"mytable"},"schema":{"fields":[{"name":"id","type":"INTEGER"},{"name":"name","type":"STRING"}]}}`
	doRequest(t, th.Server, http.MethodPost, "/bigquery/v2/projects/test-project/datasets/testds/tables", strings.NewReader(tblBody)).Body.Close()

	insBody := `{"rows":[{"f":[{"v":"1"},{"v":"hello"}]},{"f":[{"v":"2"},{"v":"world"}]}]}`
	doRequest(t, th.Server, http.MethodPost, "/bigquery/v2/projects/test-project/datasets/testds/tables/mytable:insertAll", strings.NewReader(insBody)).Body.Close()

	qBody := `{"query":"SELECT name FROM test-project:testds.mytable WHERE id = 1 LIMIT 10"}`
	resp := doRequest(t, th.Server, http.MethodPost, "/bigquery/v2/projects/test-project/queries", strings.NewReader(qBody))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("BQ query: expected 200, got %d", resp.StatusCode)
	}
}

func TestCloudTasksDispatch(t *testing.T) {
	ctb := tasks.NewMemoryCloudTasksBackend()
	cfg := router.RouterConfig{CloudTasks: ctb, DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	doRequest(t, th.Server, http.MethodPost, "/v2/projects/test-project/locations/us-central1/queues?queueId=ct-test", strings.NewReader("{}")).Body.Close()

	taskBody := `{"task":{"httpRequest":{"url":"http://localhost:9999/ct-test","httpMethod":"POST","body":"dGVzdA=="}}}`
	resp := doRequest(t, th.Server, http.MethodPost, "/v2/projects/test-project/locations/us-central1/queues/ct-test/tasks", strings.NewReader(taskBody))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("CT task create: expected 200, got %d", resp.StatusCode)
	}

	pauseResp := doRequest(t, th.Server, http.MethodPost, "/v2/projects/test-project/locations/us-central1/queues/ct-test:pause", strings.NewReader("{}"))
	defer pauseResp.Body.Close()
	if pauseResp.StatusCode != http.StatusOK {
		t.Errorf("CT pause: expected 200, got %d", pauseResp.StatusCode)
	}
}

func TestLoggingWriteRead(t *testing.T) {
	lb := logging.NewMemoryLoggingBackend()
	cfg := router.RouterConfig{Logging: lb, DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	body := `{"entries":[{"logName":"projects/test-project/logs/app","severity":"INFO","textPayload":"test message"}]}`
	resp := doRequest(t, th.Server, http.MethodPost, "/v2/entries:write", strings.NewReader(body))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Logging write: expected 200, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, th.Server, http.MethodGet, "/-/logs", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Logging list: expected 200, got %d", resp2.StatusCode)
	}
}

func TestMonitoringIngestQuery(t *testing.T) {
	mb := monitoring.NewMemoryMonitoringBackend()
	cfg := router.RouterConfig{Monitoring: mb, DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	body := `{"timeSeries":[{"metric":{"type":"custom.googleapis.com/test"},"resource":{"type":"global"},"points":[{"interval":{"endTime":"2026-01-01T00:00:00Z"},"value":{"int64Value":42}}]}]}`
	resp := doRequest(t, th.Server, http.MethodPost, "/v3/projects/test-project/timeSeries:createTimeSeries", strings.NewReader(body))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Monitoring write: expected 200, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, th.Server, http.MethodGet, "/v3/projects/test-project/timeSeries?filter=metric.type=\"custom.googleapis.com/test\"", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Monitoring query: expected 200, got %d", resp2.StatusCode)
	}
}

func TestErrorReporting(t *testing.T) {
	cfg := router.RouterConfig{DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	body := `{"message":"test error","serviceContext":{"service":"my-svc","version":"1.0"}}`
	resp := doRequest(t, th.Server, http.MethodPost, "/v1beta1/projects/test-project/events:report", strings.NewReader(body))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Error report: expected 200, got %d", resp.StatusCode)
	}

	resp2 := doRequest(t, th.Server, http.MethodGet, "/-/errors", nil)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Error list: expected 200, got %d", resp2.StatusCode)
	}
}

func TestKMSEncryptDecrypt(t *testing.T) {
	km := kms.NewMemoryKMSBackend()
	cfg := router.RouterConfig{KMS: km, DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	doRequest(t, th.Server, http.MethodPost, "/v1/projects/test-project/locations/global/keyRings?keyRingId=kms-test", strings.NewReader("{}")).Body.Close()
	doRequest(t, th.Server, http.MethodPost, "/v1/projects/test-project/locations/global/keyRings/kms-test/cryptoKeys?cryptoKeyId=key1", strings.NewReader(`{"purpose":"ENCRYPT_DECRYPT"}`)).Body.Close()

	encBody := `{"plaintext":"aGVsbG8="}`
	resp := doRequest(t, th.Server, http.MethodPost, "/v1/projects/test-project/locations/global/keyRings/kms-test/cryptoKeys/key1:encrypt", strings.NewReader(encBody))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("KMS encrypt: expected 200, got %d", resp.StatusCode)
	}
}

func TestPubSubDLQ(t *testing.T) {
	psb := pubsub.NewMemoryPubSubBackend()
	cfg := router.RouterConfig{PubSub: psb, DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	doRequest(t, th.Server, http.MethodPut, "/v1/projects/test-project/topics/d-main", strings.NewReader(`{"name":"projects/test-project/topics/d-main"}`)).Body.Close()
	doRequest(t, th.Server, http.MethodPut, "/v1/projects/test-project/topics/d-dlq", strings.NewReader(`{"name":"projects/test-project/topics/d-dlq"}`)).Body.Close()
	doRequest(t, th.Server, http.MethodPut, "/v1/projects/test-project/subscriptions/d-dlq-sub", strings.NewReader(`{"name":"projects/test-project/subscriptions/d-dlq-sub","topic":"projects/test-project/topics/d-dlq"}`)).Body.Close()

	subBody := `{"name":"projects/test-project/subscriptions/d-main-sub","topic":"projects/test-project/topics/d-main","ackDeadlineSeconds":1,"deadLetterPolicy":{"deadLetterTopic":"projects/test-project/topics/d-dlq","maxDeliveryAttempts":2}}`
	doRequest(t, th.Server, http.MethodPut, "/v1/projects/test-project/subscriptions/d-main-sub", strings.NewReader(subBody)).Body.Close()

	psb.Publish(nil, "test-project", "d-main", []*pubsubmodel.PubSubMessage{{Data: []byte("dlq")}})
	time.Sleep(1100 * time.Millisecond)
	psb.Pull(nil, "test-project", "d-main-sub", 10)
	time.Sleep(1100 * time.Millisecond)
	psb.Pull(nil, "test-project", "d-main-sub", 10)

	msgs, _ := psb.Pull(nil, "test-project", "d-dlq-sub", 10)
	if len(msgs) != 1 {
		t.Errorf("DLQ: expected 1 message, got %d", len(msgs))
	}
}

func TestPubSubFilter(t *testing.T) {
	psb := pubsub.NewMemoryPubSubBackend()
	cfg := router.RouterConfig{PubSub: psb, DefaultProject: "test-project"}
	th := newEmulator(t, cfg)
	defer th.Server.Close()

	doRequest(t, th.Server, http.MethodPut, "/v1/projects/test-project/topics/f-topic", strings.NewReader(`{"name":"projects/test-project/topics/f-topic"}`)).Body.Close()
	doRequest(t, th.Server, http.MethodPut, "/v1/projects/test-project/subscriptions/f-sub", strings.NewReader(`{"name":"projects/test-project/subscriptions/f-sub","topic":"projects/test-project/topics/f-topic","filter":"attributes.env = \"prod\""}`)).Body.Close()

	psb.Publish(nil, "test-project", "f-topic", []*pubsubmodel.PubSubMessage{
		{Data: []byte("p"), Attributes: map[string]string{"env": "prod"}},
		{Data: []byte("d"), Attributes: map[string]string{"env": "dev"}},
	})
	msgs, _ := psb.Pull(nil, "test-project", "f-sub", 10)
	if len(msgs) != 1 {
		t.Errorf("Filter: expected 1 message, got %d", len(msgs))
	}
}

var _ = kms.NewMemoryKMSBackend
