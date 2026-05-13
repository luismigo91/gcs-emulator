package api_test

import (
	"net/http/httptest"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	secretmanager "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	tasks "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/backend"
	kms "github.com/luismiguelgilolivert/gcs-emulator/internal/kms/backend"
	logging "github.com/luismiguelgilolivert/gcs-emulator/internal/logging/backend"
	monitoring "github.com/luismiguelgilolivert/gcs-emulator/internal/monitoring/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/router"
)

type TestHarness struct {
	Server         *httptest.Server
	GCS            backend.Backend
	PubSub         pubsub.PubSubBackend
	SecretManager  secretmanager.SecretManagerBackend
	CloudTasks     tasks.CloudTasksBackend
	KMS            kms.KMSBackend
	Logging        logging.LoggingBackend
	Monitoring     monitoring.MonitoringBackend
}

func newEmulator(t interface{ Helper() }, cfg router.RouterConfig) *TestHarness {
	t.Helper()
	mux := router.NewWithConfig(cfg)
	srv := httptest.NewServer(mux)
	return &TestHarness{
		Server:        srv,
		GCS:           cfg.Backend,
		PubSub:        cfg.PubSub,
		SecretManager: cfg.SecretManager,
		CloudTasks:    cfg.CloudTasks,
		KMS:           cfg.KMS,
		Logging:       cfg.Logging,
		Monitoring:    cfg.Monitoring,
	}
}

var _ = backend.NewMemoryBackend
