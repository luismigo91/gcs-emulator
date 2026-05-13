package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/model"
	pubsub "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/backend"
	pubsubmodel "github.com/luismiguelgilolivert/gcs-emulator/internal/pubsub/model"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/util"
)

type Handler struct {
	Backend         backend.Backend
	DefaultProject  string
	PubSub          pubsub.PubSubBackend
	HasCloudTasks   bool
	HasKMS          bool
	HasLogging      bool
	HasMonitoring   bool
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	services := map[string]string{"gcs": "available"}
	if h.PubSub != nil {
		services["pubsub"] = "available"
	}
	services["secretmanager"] = "available"
	if h.HasCloudTasks {
		services["cloudtasks"] = "available"
	}
	if h.HasKMS {
		services["kms"] = "available"
	}
	if h.HasLogging {
		services["logging"] = "available"
	}
	if h.HasMonitoring {
		services["monitoring"] = "available"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"services": services,
	})
}

func (h *Handler) IAMHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")

	switch r.Method {
	case http.MethodGet:
		h.getBucketIAMPolicy(w, r, bucketName)
	case http.MethodPost:
		h.setBucketIAMPolicy(w, r, bucketName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) IAMTestPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")
	h.testBucketIAMPermissions(w, r, bucketName)
}

func (h *Handler) NotificationListHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")

	switch r.Method {
	case http.MethodGet:
		h.listNotifications(w, r, bucketName)
	case http.MethodPost:
		h.createNotification(w, r, bucketName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) NotificationHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")
	notificationID := r.PathValue("notification")

	switch r.Method {
	case http.MethodGet:
		h.getNotification(w, r, bucketName, notificationID)
	case http.MethodDelete:
		h.deleteNotification(w, r, bucketName, notificationID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) LifecycleHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")

	switch r.Method {
	case http.MethodGet:
		h.getLifecycle(w, r, bucketName)
	case http.MethodPatch:
		h.setLifecycle(w, r, bucketName)
	case http.MethodDelete:
		h.deleteLifecycle(w, r, bucketName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) CORSHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")

	switch r.Method {
	case http.MethodGet:
		h.getCORS(w, r, bucketName)
	case http.MethodPatch:
		h.setCORS(w, r, bucketName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) BucketACLHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")
	entity := r.PathValue("entity")

	switch r.Method {
	case http.MethodGet:
		if entity != "" {
			h.getBucketACL(w, r, bucketName, entity)
		} else {
			h.listBucketACL(w, r, bucketName)
		}
	case http.MethodPost:
		h.createBucketACL(w, r, bucketName)
	case http.MethodPut:
		h.updateBucketACL(w, r, bucketName, entity)
	case http.MethodPatch:
		h.updateBucketACL(w, r, bucketName, entity)
	case http.MethodDelete:
		h.deleteBucketACL(w, r, bucketName, entity)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) ObjectACLHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")
	objectName := r.PathValue("object")
	entity := r.PathValue("entity")

	switch r.Method {
	case http.MethodGet:
		if entity != "" {
			h.getObjectACL(w, r, bucketName, objectName, entity)
		} else {
			h.listObjectACL(w, r, bucketName, objectName)
		}
	case http.MethodPost:
		h.createObjectACL(w, r, bucketName, objectName)
	case http.MethodPut:
		h.updateObjectACL(w, r, bucketName, objectName, entity)
	case http.MethodPatch:
		h.updateObjectACL(w, r, bucketName, objectName, entity)
	case http.MethodDelete:
		h.deleteObjectACL(w, r, bucketName, objectName, entity)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) listNotifications(w http.ResponseWriter, r *http.Request, bucket string) {
	notifications, err := h.Backend.ListNotifications(r.Context(), bucket)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if notifications == nil {
		notifications = []*model.Notification{}
	}

	response := map[string]interface{}{
		"kind":  "storage#notifications",
		"items": notifications,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) createNotification(w http.ResponseWriter, r *http.Request, bucket string) {
	var notification model.Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	created, err := h.Backend.CreateNotification(r.Context(), bucket, &notification)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, created)
}

func (h *Handler) getNotification(w http.ResponseWriter, r *http.Request, bucket, id string) {
	notification, err := h.Backend.GetNotification(r.Context(), bucket, id)
	if err != nil {
		if err == backend.ErrBucketNotFound || err == backend.ErrNotificationNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, notification)
}

func (h *Handler) deleteNotification(w http.ResponseWriter, r *http.Request, bucket, id string) {
	err := h.Backend.DeleteNotification(r.Context(), bucket, id)
	if err != nil {
		if err == backend.ErrBucketNotFound || err == backend.ErrNotificationNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) BucketHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")

	switch r.Method {
	case http.MethodPost:
		if bucketName == "" {
			h.createBucket(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	case http.MethodGet:
		if bucketName == "" {
			h.listBuckets(w, r)
		} else {
			h.getBucket(w, r, bucketName)
		}
	case http.MethodHead:
		if bucketName != "" {
			h.headBucket(w, r, bucketName)
		}
	case http.MethodDelete:
		if bucketName != "" {
			h.deleteBucket(w, r, bucketName)
		}
	case http.MethodPatch:
		if bucketName != "" {
			h.updateBucket(w, r, bucketName)
		}
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) ObjectListHandler(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")

	switch r.Method {
	case http.MethodGet:
		h.listObjects(w, r, bucket)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) ObjectHandler(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	object := r.PathValue("object")

	switch r.Method {
	case http.MethodPost:
		if strings.HasSuffix(object, "/compose") {
			destName := strings.TrimSuffix(object, "/compose")
			h.composeObjects(w, r, bucket, destName)
		} else if strings.Contains(object, "/copyTo/") {
			h.copyObject(w, r, bucket, object)
		} else if strings.Contains(object, "/rewriteTo/") {
			h.rewriteObject(w, r, bucket, object)
		} else {
			h.createObject(w, r, bucket)
		}
	case http.MethodGet:
		if r.URL.Query().Get("alt") == "media" {
			h.downloadObject(w, r, bucket, object)
		} else {
			h.getObject(w, r, bucket, object)
		}
	case http.MethodHead:
		h.headObject(w, r, bucket, object)
	case http.MethodDelete:
		h.deleteObject(w, r, bucket, object)
	case http.MethodPatch:
		h.updateObject(w, r, bucket, object)
	case http.MethodPut:
		h.updateObject(w, r, bucket, object)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err == nil && mediaType == "multipart/related" {
		h.multipartUpload(w, r, bucket)
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.createObject(w, r, bucket)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) ResumableUploadHandler(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	object := r.PathValue("object")

	uploadCommand := r.Header.Get("X-Goog-Upload-Command")
	uploadURL := r.URL.Query().Get("upload_id")

	switch {
	case uploadCommand == "start" || uploadCommand == "":
		h.createResumableUpload(w, r, bucket, object)
	case strings.Contains(uploadCommand, "finalize"):
		h.finalizeResumableUpload(w, r, bucket, object)
	case uploadURL != "" && r.Method == http.MethodPut:
		h.uploadChunk(w, r, uploadURL, bucket, object)
	case uploadCommand == "query":
		h.queryResumableUpload(w, r, uploadURL)
	default:
		if r.Method == http.MethodPut {
			h.uploadChunk(w, r, uploadURL, bucket, object)
		} else {
			writeError(w, http.StatusBadRequest, "Invalid upload command")
		}
	}
}

func (h *Handler) XMLAPIHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	bucket := parts[0]
	object := ""
	if len(parts) > 1 {
		object = parts[1]
	}

	switch r.Method {
	case http.MethodPut:
		if object != "" {
			h.xmlUploadObject(w, r, bucket, object)
		} else {
			h.xmlCreateBucket(w, r, bucket)
		}
	case http.MethodGet:
		if object != "" {
			h.xmlDownloadObject(w, r, bucket, object)
		} else {
			h.xmlListObjects(w, r, bucket)
		}
	case http.MethodDelete:
		if object != "" {
			h.xmlDeleteObject(w, r, bucket, object)
		} else {
			h.xmlDeleteBucket(w, r, bucket)
		}
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) listBuckets(w http.ResponseWriter, r *http.Request) {
	project := util.GetProjectID(r)
	query := r.URL.Query()
	projectFilter := query.Get("project")
	if projectFilter == "" {
		projectFilter = project
	}

	buckets, err := h.Backend.ListBuckets(r.Context(), backend.ListBucketsParams{
		Project: projectFilter,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if buckets == nil {
		buckets = []*model.Bucket{}
	}

	response := map[string]interface{}{
		"kind":     "storage#buckets",
		"items":    buckets,
		"prefixes": []string{},
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) createBucket(w http.ResponseWriter, r *http.Request) {
	var bucket model.Bucket
	if err := json.NewDecoder(r.Body).Decode(&bucket); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if bucket.Name == "" {
		writeError(w, http.StatusBadRequest, "Bucket name is required")
		return
	}

	bucket.Kind = "storage#bucket"
	bucket.ProjectID = util.GetProjectID(r)
	bucket.ProjectNumber = util.GenerateProjectNumber(bucket.ProjectID)
	if bucket.Metageneration == 0 {
		bucket.Metageneration = 1
	}
	if bucket.TimeCreated.IsZero() {
		bucket.TimeCreated = time.Now()
		bucket.Updated = time.Now()
	}

	conds := parseBucketConditions(r)
	if err := h.Backend.CreateBucket(r.Context(), &bucket, conds); err != nil {
		if err == backend.ErrBucketAlreadyExists {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if err == backend.ErrPreconditionFailed {
			writeError(w, http.StatusPreconditionFailed, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSONWithHeaders(w, http.StatusOK, bucket, nil)
}

func (h *Handler) getBucket(w http.ResponseWriter, r *http.Request, name string) {
	bucket, err := h.Backend.GetBucket(r.Context(), name)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bucket)
}

func (h *Handler) headBucket(w http.ResponseWriter, r *http.Request, name string) {
	bucket, err := h.Backend.GetBucket(r.Context(), name)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", bucket.Name)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) deleteBucket(w http.ResponseWriter, r *http.Request, name string) {
	conds := parseBucketConditions(r)
	if err := h.Backend.DeleteBucket(r.Context(), name, conds); err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == backend.ErrBucketNotEmpty {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if err == backend.ErrPreconditionFailed {
			writeError(w, http.StatusPreconditionFailed, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateBucket(w http.ResponseWriter, r *http.Request, name string) {
	var attrs model.BucketUpdateAttrs
	if err := json.NewDecoder(r.Body).Decode(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	conds := parseBucketConditions(r)
	bucket, err := h.Backend.UpdateBucket(r.Context(), name, &attrs, conds)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == backend.ErrPreconditionFailed {
			writeError(w, http.StatusPreconditionFailed, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, bucket)
}

func (h *Handler) createObject(w http.ResponseWriter, r *http.Request, bucket string) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = r.PathValue("object")
	}

	if name == "" {
		writeError(w, http.StatusBadRequest, "Object name is required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = util.DetectContentTypeFromName(name)
	}

	object := &model.Object{
		Name:        name,
		Bucket:      bucket,
		Kind:        "storage#object",
		ContentType: contentType,
	}

	conds := parseObjectConditions(r)
	if err := h.Backend.CreateObject(r.Context(), bucket, object, r.Body, conds); err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == backend.ErrPreconditionFailed {
			writeError(w, http.StatusPreconditionFailed, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	obj, _ := h.Backend.GetObject(r.Context(), bucket, name, object.Generation)
	h.triggerNotification(r, bucket, name, "OBJECT_FINALIZE")
	if obj != nil {
		writeJSONWithHeaders(w, http.StatusOK, obj, nil)
	} else {
		writeJSONWithHeaders(w, http.StatusOK, object, nil)
	}
}

func (h *Handler) multipartUpload(w http.ResponseWriter, r *http.Request, bucket string) {
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid Content-Type")
		return
	}

	boundary := params["boundary"]
	if boundary == "" {
		writeError(w, http.StatusBadRequest, "Missing boundary")
		return
	}

	reader := multipart.NewReader(r.Body, boundary)

	jsonPart, err := reader.NextPart()
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing metadata part")
		return
	}

	var object model.Object
	if err := json.NewDecoder(jsonPart).Decode(&object); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid metadata JSON")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		name = object.Name
	}
	if name == "" {
		name = r.PathValue("object")
	}
	object.Name = name
	object.Bucket = bucket
	object.Kind = "storage#object"

	contentPart, err := reader.NextPart()
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing content part")
		return
	}

	conds := parseObjectConditions(r)
	if err := h.Backend.CreateObject(r.Context(), bucket, &object, contentPart, conds); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	obj, _ := h.Backend.GetObject(r.Context(), bucket, name, object.Generation)
	if obj != nil {
		writeJSONWithHeaders(w, http.StatusOK, obj, nil)
	} else {
		writeJSONWithHeaders(w, http.StatusOK, &object, nil)
	}
}

func (h *Handler) listObjects(w http.ResponseWriter, r *http.Request, bucket string) {
	query := r.URL.Query()
	maxResults, _ := strconv.Atoi(query.Get("maxResults"))

	params := backend.ListObjectsParams{
		Prefix:                   query.Get("prefix"),
		Delimiter:                query.Get("delimiter"),
		Versions:                 query.Get("versions") == "true",
		StartOffset:              query.Get("startOffset"),
		EndOffset:                query.Get("endOffset"),
		IncludeTrailingDelimiter: query.Get("includeTrailingDelimiter") == "true",
		MaxResults:               maxResults,
		PageToken:                query.Get("pageToken"),
	}

	resp, err := h.Backend.ListObjects(r.Context(), bucket, params)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"kind":     "storage#objects",
		"items":    resp.Items,
		"prefixes": resp.Prefixes,
	}
	if resp.NextPageToken != "" {
		response["nextPageToken"] = resp.NextPageToken
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) getObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	generation, _ := util.ParseInt64Param(r.URL.Query().Get("generation"), "generation")

	obj, err := h.Backend.GetObject(r.Context(), bucket, object, generation)
	if err != nil {
		if err == backend.ErrObjectNotFound || err == backend.ErrGenerationNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")
	writeJSON(w, http.StatusOK, obj)
}

func (h *Handler) headObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	generation, _ := util.ParseInt64Param(r.URL.Query().Get("generation"), "generation")

	obj, err := h.Backend.GetObject(r.Context(), bucket, object, generation)
	if err != nil {
		if err == backend.ErrObjectNotFound || err == backend.ErrGenerationNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.Size, 10))
	w.Header().Set("X-Goog-Generation", strconv.FormatInt(obj.Generation, 10))
	w.Header().Set("X-Goog-Hash", fmt.Sprintf("crc32c=%s,md5=%s", obj.CRC32C, obj.MD5Hash))
	w.Header().Set("ETag", fmt.Sprintf("%q", util.GenerateETag(obj.Generation)))
	w.Header().Set("Last-Modified", obj.Updated.Format(http.TimeFormat))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) downloadObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	generation, _ := util.ParseInt64Param(r.URL.Query().Get("generation"), "generation")

	obj, err := h.Backend.GetObject(r.Context(), bucket, object, generation)
	if err != nil {
		if err == backend.ErrObjectNotFound || err == backend.ErrGenerationNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	content, err := h.Backend.GetObjectContent(r.Context(), bucket, object, generation)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	data, err := io.ReadAll(content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.serveObjectContent(w, r, obj, data)
}

func (h *Handler) serveObjectContent(w http.ResponseWriter, r *http.Request, obj *model.Object, data []byte) {
	status := http.StatusOK
	content := io.Reader(bytes.NewReader(data))
	contentLength := int64(len(data))

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		start, end, satisfiable := parseRange(rangeHeader, contentLength)
		if satisfiable {
			status = http.StatusPartialContent
			content = bytes.NewReader(data[start : end+1])
			contentLength = end - start + 1
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, contentLength))
		} else if start >= contentLength {
			status = http.StatusRequestedRangeNotSatisfiable
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", contentLength))
			content = bytes.NewReader([]byte(fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?><Error><Code>InvalidRange</Code><Message>The requested range cannot be satisfied.</Message><Details>%s</Details></Error>`, rangeHeader)))
			w.Header().Set("Content-Type", "application/xml; charset=UTF-8")
		}
	}

	if obj.ContentType != "" {
		w.Header().Set("Content-Type", obj.ContentType)
	}
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("X-Goog-Generation", strconv.FormatInt(obj.Generation, 10))
	w.Header().Set("X-Goog-Hash", fmt.Sprintf("crc32c=%s,md5=%s", obj.CRC32C, obj.MD5Hash))
	w.Header().Set("X-Goog-Stored-Content-Length", strconv.FormatInt(obj.Size, 10))
	w.Header().Set("Last-Modified", obj.Updated.Format(http.TimeFormat))
	w.Header().Set("ETag", fmt.Sprintf("%q", util.GenerateETag(obj.Generation)))
	for name, value := range obj.Metadata {
		w.Header().Set("X-Goog-Meta-"+name, value)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if obj.ContentEncoding != "" {
		w.Header().Set("Content-Encoding", obj.ContentEncoding)
	}
	if obj.CacheControl != "" {
		w.Header().Set("Cache-Control", obj.CacheControl)
	}
	if obj.ContentDisposition != "" {
		w.Header().Set("Content-Disposition", obj.ContentDisposition)
	}
	if obj.ContentLanguage != "" {
		w.Header().Set("Content-Language", obj.ContentLanguage)
	}

	w.WriteHeader(status)
	if r.Method == http.MethodGet {
		io.Copy(w, content)
	}
}

func (h *Handler) deleteObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	generation, _ := util.ParseInt64Param(r.URL.Query().Get("generation"), "generation")
	conds := parseObjectConditions(r)

	if err := h.Backend.DeleteObject(r.Context(), bucket, object, generation, conds); err != nil {
		if err == backend.ErrObjectNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == backend.ErrPreconditionFailed {
			writeError(w, http.StatusPreconditionFailed, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.triggerNotification(r, bucket, object, "OBJECT_DELETE")

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	var attrs model.ObjectUpdateAttrs
	if err := json.NewDecoder(r.Body).Decode(&attrs); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	conds := parseObjectConditions(r)
	obj, err := h.Backend.UpdateObject(r.Context(), bucket, object, &attrs, conds)
	if err != nil {
		if err == backend.ErrObjectNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if err == backend.ErrPreconditionFailed {
			writeError(w, http.StatusPreconditionFailed, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, obj)
}

func (h *Handler) composeObjects(w http.ResponseWriter, r *http.Request, bucket, destName string) {
	var req struct {
		SourceObjects []struct {
			Name string `json:"name"`
		} `json:"sourceObjects"`
		Destination *model.Object `json:"destination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sourceNames := make([]string, len(req.SourceObjects))
	for i, src := range req.SourceObjects {
		sourceNames[i] = src.Name
	}

	obj, err := h.Backend.ComposeObjects(r.Context(), bucket, destName, sourceNames, req.Destination)
	if err != nil {
		if err == backend.ErrComposeTooManySources {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSONWithHeaders(w, http.StatusOK, obj, nil)
}

func (h *Handler) copyObject(w http.ResponseWriter, r *http.Request, srcBucket, srcName string) {
	pathParts := strings.Split(srcName, "/")
	var destBucket, destName string
	for i, part := range pathParts {
		if part == "copyTo" && i+3 < len(pathParts) {
			if pathParts[i+1] == "b" && i+4 < len(pathParts) {
				destBucket = pathParts[i+2]
				destName = strings.Join(pathParts[i+4:], "/")
			} else {
				destBucket = pathParts[i+1]
				destName = strings.Join(pathParts[i+3:], "/")
			}
			srcName = strings.Join(pathParts[:i], "/")
			break
		}
	}

	if destBucket == "" || destName == "" {
		writeError(w, http.StatusBadRequest, "Invalid copy path")
		return
	}

	conds := parseObjectConditions(r)
	obj, err := h.Backend.CopyObject(r.Context(), srcBucket, srcName, destBucket, destName, conds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSONWithHeaders(w, http.StatusOK, obj, nil)
}

func (h *Handler) rewriteObject(w http.ResponseWriter, r *http.Request, srcBucket, srcName string) {
	pathParts := strings.Split(srcName, "/")
	var destBucket, destName string
	for i, part := range pathParts {
		if part == "rewriteTo" && i+3 < len(pathParts) {
			if pathParts[i+1] == "b" && i+4 < len(pathParts) {
				destBucket = pathParts[i+2]
				destName = strings.Join(pathParts[i+4:], "/")
			} else {
				destBucket = pathParts[i+1]
				destName = strings.Join(pathParts[i+3:], "/")
			}
			srcName = strings.Join(pathParts[:i], "/")
			break
		}
	}

	if destBucket == "" || destName == "" {
		writeError(w, http.StatusBadRequest, "Invalid rewrite path")
		return
	}

	token := r.URL.Query().Get("rewriteToken")
	conds := parseObjectConditions(r)
	obj, newToken, err := h.Backend.RewriteObject(r.Context(), srcBucket, srcName, destBucket, destName, token, conds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"kind":                  "storage#rewriteResponse",
		"totalBytesRewritten":   fmt.Sprintf("%d", obj.Size),
		"objectSize":            fmt.Sprintf("%d", obj.Size),
		"done":                  true,
		"resource":              obj,
		"rewriteToken":          newToken,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) createResumableUpload(w http.ResponseWriter, r *http.Request, bucket, object string) {
	sessionID := backend.GenerateID()
	uploadURL := r.URL.Path + "?upload_id=" + sessionID

	name := r.URL.Query().Get("name")
	if name == "" {
		name = object
	}

	var metadata map[string]interface{}
	if r.ContentLength > 0 {
		json.NewDecoder(r.Body).Decode(&metadata)
	}

	session := &backend.UploadSession{
		ID:         sessionID,
		Bucket:     bucket,
		ObjectName: name,
		Metadata:   metadata,
		Chunks:     make(map[int64][]byte),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}
	backend.DefaultSessionStore.Create(session)

	w.Header().Set("X-Goog-Upload-URL", uploadURL)
	w.Header().Set("X-Goog-Upload-Status", "active")
	w.Header().Set("X-Goog-Upload-Command", "start")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) uploadChunk(w http.ResponseWriter, r *http.Request, uploadURL, bucket, object string) {
	sessionID := r.URL.Query().Get("upload_id")
	if sessionID == "" {
		parts := strings.Split(uploadURL, "upload_id=")
		if len(parts) > 1 {
			sessionID = parts[1]
		}
	}

	session, exists := backend.DefaultSessionStore.Get(sessionID)
	if !exists {
		writeError(w, http.StatusNotFound, "Upload session not found")
		return
	}

	contentRange := r.Header.Get("Content-Range")
	if contentRange == "" {
		writeError(w, http.StatusBadRequest, "Content-Range header required")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	start, _, _ := parseContentRange(contentRange)
	backend.DefaultSessionStore.StoreChunk(sessionID, start, data)

	w.Header().Set("X-Goog-Upload-Status", "active")
	w.Header().Set("X-Goog-Upload-Offset", strconv.FormatInt(session.Offset, 10))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) finalizeResumableUpload(w http.ResponseWriter, r *http.Request, bucket, object string) {
	sessionID := r.URL.Query().Get("upload_id")

	session, exists := backend.DefaultSessionStore.Get(sessionID)
	if !exists {
		writeError(w, http.StatusNotFound, "Upload session not found")
		return
	}

	if r.ContentLength > 0 {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		contentRange := r.Header.Get("Content-Range")
		if contentRange != "" {
			start, _, _ := parseContentRange(contentRange)
			backend.DefaultSessionStore.StoreChunk(sessionID, start, data)
		}
	}

	data := backend.DefaultSessionStore.MergeChunks(sessionID)
	backend.DefaultSessionStore.Delete(sessionID)

	name := r.URL.Query().Get("name")
	if name == "" {
		name = session.ObjectName
	}

	obj := &model.Object{
		Name:   name,
		Bucket: bucket,
		Kind:   "storage#object",
	}

	if err := h.Backend.CreateObject(r.Context(), bucket, obj, bytes.NewReader(data), backend.Conditions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("X-Goog-Upload-Status", "final")
	writeJSONWithHeaders(w, http.StatusOK, obj, nil)
}

func (h *Handler) queryResumableUpload(w http.ResponseWriter, r *http.Request, uploadURL string) {
	sessionID := r.URL.Query().Get("upload_id")

	session, exists := backend.DefaultSessionStore.Get(sessionID)
	if !exists {
		writeError(w, http.StatusNotFound, "Upload session not found")
		return
	}

	w.Header().Set("X-Goog-Upload-Status", "active")
	w.Header().Set("X-Goog-Upload-Offset", strconv.FormatInt(session.Offset, 10))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) xmlUploadObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	obj := &model.Object{
		Name:   object,
		Bucket: bucket,
	}

	if err := h.Backend.CreateObject(r.Context(), bucket, obj, r.Body, backend.Conditions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) xmlDownloadObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	obj, err := h.Backend.GetObject(r.Context(), bucket, object, 0)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	content, err := h.Backend.GetObjectContent(r.Context(), bucket, object, 0)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	data, err := io.ReadAll(content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.serveObjectContent(w, r, obj, data)
}

func (h *Handler) xmlListObjects(w http.ResponseWriter, r *http.Request, bucket string) {
	params := backend.ListObjectsParams{
		Prefix:    r.URL.Query().Get("prefix"),
		Delimiter: r.URL.Query().Get("delimiter"),
	}

	resp, err := h.Backend.ListObjects(r.Context(), bucket, params)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	w.Write([]byte(`<ListBucketResult>`))
	w.Write([]byte(`<Name>` + bucket + `</Name>`))
	w.Write([]byte(`<Prefix>` + params.Prefix + `</Prefix>`))
	w.Write([]byte(`<Delimiter>` + params.Delimiter + `</Delimiter>`))
	w.Write([]byte(`<KeyCount>` + strconv.Itoa(len(resp.Items)) + `</KeyCount>`))
	for _, prefix := range resp.Prefixes {
		w.Write([]byte(`<CommonPrefixes><Prefix>` + prefix + `</Prefix></CommonPrefixes>`))
	}
	for _, obj := range resp.Items {
		w.Write([]byte(`<Contents>`))
		w.Write([]byte(`<Key>` + obj.Name + `</Key>`))
		w.Write([]byte(`<Size>` + strconv.FormatInt(obj.Size, 10) + `</Size>`))
		w.Write([]byte(`<LastModified>` + obj.Updated.Format(time.RFC3339) + `</LastModified>`))
		w.Write([]byte(`</Contents>`))
	}
	w.Write([]byte(`</ListBucketResult>`))
}

func (h *Handler) xmlCreateBucket(w http.ResponseWriter, r *http.Request, bucket string) {
	b := &model.Bucket{
		Name: bucket,
		Kind: "storage#bucket",
	}

	if err := h.Backend.CreateBucket(r.Context(), b, backend.Conditions{}); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) xmlDeleteObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	if err := h.Backend.DeleteObject(r.Context(), bucket, object, 0, backend.Conditions{}); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) xmlDeleteBucket(w http.ResponseWriter, r *http.Request, bucket string) {
	if err := h.Backend.DeleteBucket(r.Context(), bucket, backend.Conditions{}); err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
		} else if err == backend.ErrBucketNotEmpty {
			writeError(w, http.StatusConflict, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseRange(rangeHeaderValue string, contentLength int64) (start int64, end int64, satisfiable bool) {
	parts := strings.SplitN(rangeHeaderValue, "=", 2)
	if len(parts) != 2 || parts[0] != "bytes" {
		return 0, contentLength - 1, true
	}

	rangeSpec := parts[1]
	if len(rangeSpec) == 0 {
		return 0, contentLength - 1, true
	}

	if rangeSpec[0] == '-' {
		offsetFromEnd, err := strconv.ParseInt(rangeSpec, 10, 64)
		if err != nil {
			return 0, contentLength - 1, true
		}
		start = contentLength + offsetFromEnd
		if start < 0 {
			start = 0
		}
		return start, contentLength - 1, true
	}

	rangeParts := strings.SplitN(rangeSpec, "-", 2)
	if len(rangeParts) != 2 {
		return 0, contentLength - 1, true
	}

	var s int64
	s, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return 0, contentLength - 1, true
	}

	if rangeParts[1] == "" {
		return s, contentLength - 1, s < contentLength
	}

	var e int64
	e, err = strconv.ParseInt(rangeParts[1], 10, 64)
	if err != nil {
		return 0, contentLength - 1, true
	}

	if e >= contentLength {
		e = contentLength - 1
	}

	if s >= contentLength {
		return s, e, false
	}

	if e < s {
		return 0, contentLength - 1, true
	}

	return s, e, true
}

func parseContentRange(contentRange string) (start int64, end int64, total int64) {
	parts := strings.SplitN(contentRange, " ", 2)
	if len(parts) != 2 {
		return 0, 0, 0
	}

	rangeParts := strings.SplitN(parts[1], "/", 2)
	if len(rangeParts) != 2 {
		return 0, 0, 0
	}

	if rangeParts[1] != "*" {
		total, _ = strconv.ParseInt(rangeParts[1], 10, 64)
	}

	bytesParts := strings.SplitN(rangeParts[0], "-", 2)
	if len(bytesParts) == 2 {
		start, _ = strconv.ParseInt(bytesParts[0], 10, 64)
		end, _ = strconv.ParseInt(bytesParts[1], 10, 64)
	}

	return start, end, total
}

func parseBucketConditions(r *http.Request) backend.Conditions {
	var conds backend.Conditions

	if v := r.URL.Query().Get("ifMetagenerationMatch"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			conds.MetagenerationMatch = val
			conds.HasMetagenerationMatch = true
		}
	}
	if v := r.URL.Query().Get("ifMetagenerationNotMatch"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			conds.MetagenerationNotMatch = val
			conds.HasMetagenerationNotMatch = true
		}
	}

	return conds
}

func parseObjectConditions(r *http.Request) backend.Conditions {
	var conds backend.Conditions

	if v := r.URL.Query().Get("ifGenerationMatch"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			if val == 0 {
				conds.DoesNotExist = true
			} else {
				conds.GenerationMatch = val
				conds.HasGenerationMatch = true
			}
		}
	}
	if v := r.URL.Query().Get("ifGenerationNotMatch"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			conds.GenerationNotMatch = val
			conds.HasGenerationNotMatch = true
		}
	}
	if v := r.URL.Query().Get("ifMetagenerationMatch"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			conds.MetagenerationMatch = val
			conds.HasMetagenerationMatch = true
		}
	}
	if v := r.URL.Query().Get("ifMetagenerationNotMatch"); v != "" {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			conds.MetagenerationNotMatch = val
			conds.HasMetagenerationNotMatch = true
		}
	}

	return conds
}

func (h *Handler) getBucketIAMPolicy(w http.ResponseWriter, r *http.Request, bucket string) {
	policy, err := h.Backend.GetBucketIAMPolicy(r.Context(), bucket)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, policy)
}

func (h *Handler) setBucketIAMPolicy(w http.ResponseWriter, r *http.Request, bucket string) {
	var policy model.Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedPolicy, err := h.Backend.SetBucketIAMPolicy(r.Context(), bucket, &policy)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updatedPolicy)
}

func (h *Handler) testBucketIAMPermissions(w http.ResponseWriter, r *http.Request, bucket string) {
	permissions := r.URL.Query()["permissions"]
	if len(permissions) == 0 {
		writeError(w, http.StatusBadRequest, "permissions query parameter is required")
		return
	}

	allowed, err := h.Backend.TestBucketIAMPermissions(r.Context(), bucket, permissions)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string][]string{
		"permissions": allowed,
	})
}

func (h *Handler) getLifecycle(w http.ResponseWriter, r *http.Request, bucket string) {
	lifecycle, err := h.Backend.GetLifecycle(r.Context(), bucket)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, lifecycle)
}

func (h *Handler) setLifecycle(w http.ResponseWriter, r *http.Request, bucket string) {
	var lifecycle model.Lifecycle
	if err := json.NewDecoder(r.Body).Decode(&lifecycle); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.SetLifecycle(r.Context(), bucket, &lifecycle)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteLifecycle(w http.ResponseWriter, r *http.Request, bucket string) {
	if err := h.Backend.DeleteLifecycle(r.Context(), bucket); err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) getCORS(w http.ResponseWriter, r *http.Request, bucket string) {
	rules, err := h.Backend.GetCORS(r.Context(), bucket)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  "storage#cors",
		"items": rules,
	})
}

func (h *Handler) setCORS(w http.ResponseWriter, r *http.Request, bucket string) {
	var rules []model.CORSRule
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if rules == nil {
		rules = []model.CORSRule{}
	}
	result, err := h.Backend.SetCORS(r.Context(), bucket, rules)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  "storage#cors",
		"items": result,
	})
}

func (h *Handler) listBucketACL(w http.ResponseWriter, r *http.Request, bucket string) {
	acls, err := h.Backend.ListBucketACL(r.Context(), bucket)
	if err != nil {
		if err == backend.ErrBucketNotFound {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if acls == nil {
		acls = []*model.BucketACL{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  "storage#bucketAccessControls",
		"items": acls,
	})
}

func (h *Handler) getBucketACL(w http.ResponseWriter, r *http.Request, bucket, entity string) {
	acl, err := h.Backend.GetBucketACL(r.Context(), bucket, entity)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acl)
}

func (h *Handler) createBucketACL(w http.ResponseWriter, r *http.Request, bucket string) {
	var acl model.BucketACL
	if err := json.NewDecoder(r.Body).Decode(&acl); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.CreateBucketACL(r.Context(), bucket, &acl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) updateBucketACL(w http.ResponseWriter, r *http.Request, bucket, entity string) {
	var acl model.BucketACL
	if err := json.NewDecoder(r.Body).Decode(&acl); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.UpdateBucketACL(r.Context(), bucket, entity, &acl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteBucketACL(w http.ResponseWriter, r *http.Request, bucket, entity string) {
	if err := h.Backend.DeleteBucketACL(r.Context(), bucket, entity); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listObjectACL(w http.ResponseWriter, r *http.Request, bucket, object string) {
	acls, err := h.Backend.ListObjectACL(r.Context(), bucket, object, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if acls == nil {
		acls = []*model.ObjectACL{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  "storage#objectAccessControls",
		"items": acls,
	})
}

func (h *Handler) getObjectACL(w http.ResponseWriter, r *http.Request, bucket, object, entity string) {
	acl, err := h.Backend.GetObjectACL(r.Context(), bucket, object, entity, 0)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acl)
}

func (h *Handler) createObjectACL(w http.ResponseWriter, r *http.Request, bucket, object string) {
	var acl model.ObjectACL
	if err := json.NewDecoder(r.Body).Decode(&acl); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.CreateObjectACL(r.Context(), bucket, object, &acl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) updateObjectACL(w http.ResponseWriter, r *http.Request, bucket, object, entity string) {
	var acl model.ObjectACL
	if err := json.NewDecoder(r.Body).Decode(&acl); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.UpdateObjectACL(r.Context(), bucket, object, entity, &acl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteObjectACL(w http.ResponseWriter, r *http.Request, bucket, object, entity string) {
	if err := h.Backend.DeleteObjectACL(r.Context(), bucket, object, entity); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) WebsiteHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")
	switch r.Method {
	case http.MethodGet:
		h.getWebsite(w, r, bucketName)
	case http.MethodPut, http.MethodPatch:
		h.setWebsite(w, r, bucketName)
	case http.MethodDelete:
		h.deleteWebsite(w, r, bucketName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) getWebsite(w http.ResponseWriter, r *http.Request, bucket string) {
	config, err := h.Backend.GetWebsiteConfig(r.Context(), bucket)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (h *Handler) setWebsite(w http.ResponseWriter, r *http.Request, bucket string) {
	var config model.WebsiteConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.SetWebsiteConfig(r.Context(), bucket, &config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteWebsite(w http.ResponseWriter, r *http.Request, bucket string) {
	if err := h.Backend.DeleteWebsiteConfig(r.Context(), bucket); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) EncryptionHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")
	switch r.Method {
	case http.MethodGet:
		h.getEncryption(w, r, bucketName)
	case http.MethodPut, http.MethodPatch:
		h.setEncryption(w, r, bucketName)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) getEncryption(w http.ResponseWriter, r *http.Request, bucket string) {
	config, err := h.Backend.GetEncryptionConfig(r.Context(), bucket)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func (h *Handler) setEncryption(w http.ResponseWriter, r *http.Request, bucket string) {
	var config model.EncryptionConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.SetEncryptionConfig(r.Context(), bucket, &config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) DefaultObjectACLHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucket")
	entity := r.PathValue("entity")

	switch r.Method {
	case http.MethodGet:
		if entity != "" {
			h.getDefaultObjectACL(w, r, bucketName, entity)
		} else {
			h.listDefaultObjectACL(w, r, bucketName)
		}
	case http.MethodPost:
		h.createDefaultObjectACL(w, r, bucketName)
	case http.MethodPut, http.MethodPatch:
		h.updateDefaultObjectACL(w, r, bucketName, entity)
	case http.MethodDelete:
		h.deleteDefaultObjectACL(w, r, bucketName, entity)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) listDefaultObjectACL(w http.ResponseWriter, r *http.Request, bucket string) {
	acls, err := h.Backend.ListDefaultObjectACL(r.Context(), bucket)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if acls == nil {
		acls = []*model.ObjectACL{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  "storage#objectAccessControls",
		"items": acls,
	})
}

func (h *Handler) getDefaultObjectACL(w http.ResponseWriter, r *http.Request, bucket, entity string) {
	acls, _ := h.Backend.ListDefaultObjectACL(r.Context(), bucket)
	for _, a := range acls {
		if a.Entity == entity {
			writeJSON(w, http.StatusOK, a)
			return
		}
	}
	writeError(w, http.StatusNotFound, "ACL not found")
}

func (h *Handler) createDefaultObjectACL(w http.ResponseWriter, r *http.Request, bucket string) {
	var acl model.ObjectACL
	if err := json.NewDecoder(r.Body).Decode(&acl); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.CreateDefaultObjectACL(r.Context(), bucket, &acl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) updateDefaultObjectACL(w http.ResponseWriter, r *http.Request, bucket, entity string) {
	var acl model.ObjectACL
	if err := json.NewDecoder(r.Body).Decode(&acl); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	result, err := h.Backend.UpdateDefaultObjectACL(r.Context(), bucket, entity, &acl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) deleteDefaultObjectACL(w http.ResponseWriter, r *http.Request, bucket, entity string) {
	if err := h.Backend.DeleteDefaultObjectACL(r.Context(), bucket, entity); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) triggerNotification(r *http.Request, bucket, objectName, eventType string) {
	if h.PubSub == nil {
		return
	}

	notifications, _ := h.Backend.ListNotifications(r.Context(), bucket)
	for _, n := range notifications {
		found := false
		for _, et := range n.EventType {
			if et == eventType {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if n.ObjectNamePrefix != "" && !strings.HasPrefix(objectName, n.ObjectNamePrefix) {
			continue
		}

		attrs := map[string]string{
			"bucketId":           bucket,
			"objectId":           objectName,
			"eventType":          eventType,
			"notificationConfig": n.ID,
		}
		data := fmt.Sprintf(`{"kind":"storage#object","name":"%s","bucket":"%s"}`, objectName, bucket)

		topicParts := strings.Split(n.Topic, "/")
		if len(topicParts) < 4 {
			continue
		}

		msg := &pubsubmodel.PubSubMessage{
			Data:       []byte(data),
			Attributes: attrs,
		}
		_, err := h.PubSub.Publish(r.Context(), util.GetProjectID(r), topicParts[len(topicParts)-1], []*pubsubmodel.PubSubMessage{msg})
		if err != nil {
			fmt.Fprintf(io.Discard, "notification publish failed: %v", err)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONWithHeaders(w http.ResponseWriter, status int, data interface{}, headers http.Header) {
	w.Header().Set("Content-Type", "application/json")
	for k, v := range headers {
		w.Header()[k] = v
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.NewSimpleGCPError(status, message))
}
