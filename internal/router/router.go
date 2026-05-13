package router

import (
	"net/http"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/admin"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/api"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/auth"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/metrics"
	pubsubapi "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/api"
	pubsubbackend "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	secretapi "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/api"
	secretbackend "github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/backend"
	tasksapi "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/api"
	tasksbackend "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/backend"
	kmsapi "github.com/luismiguelgilolivert/gcs-emulator/internal/kms/api"
	kmsbackend "github.com/luismiguelgilolivert/gcs-emulator/internal/kms/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/util"
)

func New(b backend.Backend, psb pubsubbackend.PubSubBackend, smb secretbackend.SecretManagerBackend, ctb tasksbackend.CloudTasksBackend, kmb kmsbackend.KMSBackend, defaultProject string) http.Handler {
	mux := http.NewServeMux()
	h := &api.Handler{Backend: b, DefaultProject: defaultProject}
	if psb != nil {
		h.PubSub = psb
	}
	mc := metrics.NewCollector()
	mc.SetObjectCountFunc(func() int64 { return 0 })
	mc.SetBucketCountFunc(func() int64 { return 0 })
	mux.HandleFunc("/metrics", mc.Handler())

	// --- GCS routes ---
	mux.HandleFunc("/storage/v1/b", h.BucketHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}", h.BucketHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/iam", h.IAMHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/iam:testIamPermissions", h.IAMTestPermissionsHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/notificationConfigs", h.NotificationListHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/notificationConfigs/{notification}", h.NotificationHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/lifecycle", h.LifecycleHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/cors", h.CORSHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/acl", h.BucketACLHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/acl/{entity}", h.BucketACLHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/o/{object}/acl", h.ObjectACLHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/o/{object}/acl/{entity}", h.ObjectACLHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/website", h.WebsiteHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/encryption", h.EncryptionHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/defaultObjectAcl", h.DefaultObjectACLHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/defaultObjectAcl/{entity}", h.DefaultObjectACLHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/o", h.ObjectListHandler)
	mux.HandleFunc("/storage/v1/b/{bucket}/o/{object...}", h.ObjectHandler)
	mux.HandleFunc("/upload/storage/v1/b/{bucket}/o", h.UploadHandler)
	mux.HandleFunc("/resumable/upload/storage/v1/b/{bucket}/o", h.ResumableUploadHandler)
	mux.HandleFunc("/resumable/upload/storage/v1/b/{bucket}/o/{object...}", h.ResumableUploadHandler)

	// --- Pub/Sub routes ---
	if psb != nil {
		ps := &pubsubapi.Handler{Backend: psb}
		mux.HandleFunc("/v1/projects/{project}/topics", ps.TopicHandler)
		mux.HandleFunc("/v1/projects/{project}/topics/{topic}", ps.TopicHandler)
		mux.HandleFunc("/v1/projects/{project}/subscriptions", ps.SubscriptionHandler)
		mux.HandleFunc("/v1/projects/{project}/subscriptions/{subscription}", ps.SubscriptionHandler)
		mux.HandleFunc("/v1/projects/{project}/schemas", ps.SchemaHandler)
		mux.HandleFunc("/v1/projects/{project}/schemas/{schema}", ps.SchemaHandler)
	}

	// --- Secret Manager routes ---
	if smb != nil {
		sm := &secretapi.Handler{Backend: smb}
		mux.HandleFunc("/v1/projects/{project}/secrets", sm.SecretsHandler)
		mux.HandleFunc("/v1/projects/{project}/secrets/{secret}", sm.SecretsHandler)
		mux.HandleFunc("/v1/projects/{project}/secrets/{secret}/versions", sm.SecretsHandler)
		mux.HandleFunc("/v1/projects/{project}/secrets/{secret}/versions/{version}", sm.SecretsHandler)
	}

	// --- Cloud Tasks routes ---
	if ctb != nil {
		ct := &tasksapi.Handler{Backend: ctb}
		mux.Handle("/v2/projects/", ct)
		h.HasCloudTasks = true
	}

	// --- KMS routes ---
	if kmb != nil {
		km := &kmsapi.Handler{Backend: kmb}
		mux.Handle("/v1/projects/{project}/locations/{location}/keyRings", km)
		mux.Handle("/v1/projects/{project}/locations/{location}/keyRings/{keyRing}", km)
		mux.Handle("/v1/projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys", km)
		mux.Handle("/v1/projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{cryptoKey}", km)
		mux.Handle("/v1/projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{cryptoKey}/cryptoKeyVersions", km)
		mux.Handle("/v1/projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{cryptoKey}/cryptoKeyVersions/{version}", km)
		h.HasKMS = true
	}

	// --- Admin & Health ---
	mux.HandleFunc("/-/health", h.HealthHandler)

	svcList := []string{"gcs"}
	if psb != nil { svcList = append(svcList, "pubsub") }
	if smb != nil { svcList = append(svcList, "secretmanager") }
	if ctb != nil { svcList = append(svcList, "cloudtasks") }
	if kmb != nil { svcList = append(svcList, "kms") }
	mux.HandleFunc("/__/services", admin.ServicesHandler(svcList))

	mux.HandleFunc("/oauth2/v4/token", auth.TokenHandler)

	// XML API (catch-all, must be last)
	mux.HandleFunc("/", h.XMLAPIHandler)

	return mc.Middleware(corsMiddleware(projectMiddleware(mux, defaultProject)))
}

func projectMiddleware(next http.Handler, defaultProject string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.Header.Get("X-Goog-User-Project")
		if projectID == "" {
			projectID = defaultProject
		}
		ctx := util.WithProjectID(r.Context(), projectID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Goog-User-Project, X-Goog-Upload-Command, Content-Range, Range")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type, Content-Length, ETag, X-Goog-Generation, X-Goog-Metageneration, X-Goog-Upload-URL, X-Goog-Upload-Status")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetProjectID(r *http.Request) string {
	return util.GetProjectID(r)
}
