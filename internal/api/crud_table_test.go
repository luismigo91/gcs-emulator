package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	secretmanager "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	kms "github.com/luismiguelgilolivert/gcs-emulator/internal/kms/backend"
	logging "github.com/luismiguelgilolivert/gcs-emulator/internal/logging/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

type crudTestCase struct {
	name        string
	cfg         router.RouterConfig
	createPath  string
	createBody  string
	getPath     string
	listPath    string
	deletePath  string
}

func TestTableDrivenCRUD(t *testing.T) {
	

	cases := []crudTestCase{
		{
			name: "SecretManager",
			cfg: router.RouterConfig{
				SecretManager: secretmanager.NewMemorySecretManagerBackend(),
				DefaultProject: "test-project",
			},
			createPath: "/v1/projects/test-project/secrets?secretId=test-secret",
			createBody: `{"replication":{"automatic":{}}}`,
			getPath:    "/v1/projects/test-project/secrets/test-secret",
			listPath:   "/v1/projects/test-project/secrets",
			deletePath: "/v1/projects/test-project/secrets/test-secret",
		},
		{
			name: "KMS",
			cfg: router.RouterConfig{
				KMS:            kms.NewMemoryKMSBackend(),
				DefaultProject: "test-project",
			},
			createPath: "/v1/projects/test-project/locations/global/keyRings?keyRingId=test-kr",
			createBody: `{}`,
			getPath:    "/v1/projects/test-project/locations/global/keyRings/test-kr",
			listPath:   "/v1/projects/test-project/locations/global/keyRings",
			deletePath: "/v1/projects/test-project/locations/global/keyRings/test-kr",
		},
		{
			name: "Logging",
			cfg: router.RouterConfig{
				Logging:        logging.NewMemoryLoggingBackend(),
				DefaultProject: "test-project",
			},
			createPath: "/v2/entries:write",
			createBody: `{"entries":[{"logName":"projects/test-project/logs/test","severity":"INFO","textPayload":"crud test"}]}`,
			listPath:  "/-/logs",
		},
		{
			name: "PubSub",
			cfg: router.RouterConfig{
				PubSub:         pubsub.NewMemoryPubSubBackend(),
				DefaultProject: "test-project",
			},
			createPath: "/v1/projects/test-project/topics/test-topic",
			createBody: `{"name":"projects/test-project/topics/test-topic"}`,
			getPath:    "/v1/projects/test-project/topics/test-topic",
			listPath:   "/v1/projects/test-project/topics",
			deletePath: "/v1/projects/test-project/topics/test-topic",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"/create", func(t *testing.T) {
			th := newEmulator(t, tc.cfg)
			defer th.Server.Close()

			if tc.createPath != "" {
				method := http.MethodPut
				if strings.Contains(tc.createPath, "?") { method = http.MethodPost }
				if strings.Contains(tc.createPath, "entries:write") { method = http.MethodPost }
				resp := doRequest(t, th.Server, method, tc.createPath, strings.NewReader(tc.createBody))
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("%s create: expected 200, got %d", tc.name, resp.StatusCode)
				}
			}
		})

		if tc.getPath != "" {
			t.Run(tc.name+"/get", func(t *testing.T) {
				
				th := newEmulator(t, tc.cfg)
				defer th.Server.Close()

				create(t, th.Server, tc.createPath, tc.createBody)
				resp := doRequest(t, th.Server, http.MethodGet, tc.getPath, nil)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("%s get: expected 200, got %d", tc.name, resp.StatusCode)
				}
			})
		}

		t.Run(tc.name+"/list", func(t *testing.T) {
			
			th := newEmulator(t, tc.cfg)
			defer th.Server.Close()

			if tc.listPath != "" {
				resp := doRequest(t, th.Server, http.MethodGet, tc.listPath, nil)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("%s list: expected 200, got %d", tc.name, resp.StatusCode)
				}
			}
		})

		if tc.deletePath != "" {
			t.Run(tc.name+"/delete", func(t *testing.T) {
				
				th := newEmulator(t, tc.cfg)
				defer th.Server.Close()

				create(t, th.Server, tc.createPath, tc.createBody)
				resp := doRequest(t, th.Server, http.MethodDelete, tc.deletePath, nil)
				defer resp.Body.Close()
				if resp.StatusCode >= 400 {
					t.Errorf("%s delete: expected 2xx, got %d", tc.name, resp.StatusCode)
				}
			})
		}
	}
}

func create(t *testing.T, srv *httptest.Server, path, body string) {
	t.Helper()
	method := http.MethodPut
	if strings.Contains(path, "?") { method = http.MethodPost }
	if strings.Contains(path, "entries:write") { method = http.MethodPost }
	resp := doRequest(t, srv, method, path, strings.NewReader(body))
	resp.Body.Close()
}

var _ = backend.NewMemoryBackend
