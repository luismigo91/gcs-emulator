package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Collector struct {
	mu               sync.RWMutex
	requestsTotal    map[string]map[string]int64
	requestDuration  []float64
	objectsTotalFunc func() int64
	bucketsTotalFunc func() int64
}

func NewCollector() *Collector {
	return &Collector{
		requestsTotal:   make(map[string]map[string]int64),
		requestDuration: make([]float64, 0),
	}
}

func (c *Collector) SetObjectCountFunc(f func() int64) {
	c.objectsTotalFunc = f
}

func (c *Collector) SetBucketCountFunc(f func() int64) {
	c.bucketsTotalFunc = f
}

func (c *Collector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()

		c.mu.Lock()
		defer c.mu.Unlock()

		if _, ok := c.requestsTotal[r.Method]; !ok {
			c.requestsTotal[r.Method] = make(map[string]int64)
		}
		c.requestsTotal[r.Method][r.URL.Path]++
		c.requestDuration = append(c.requestDuration, duration)
	})
}

func (c *Collector) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		fmt.Fprintf(w, "# HELP gcs_emulator_requests_total Total number of HTTP requests\n")
		fmt.Fprintf(w, "# TYPE gcs_emulator_requests_total counter\n")
		for method, paths := range c.requestsTotal {
			for path, count := range paths {
				fmt.Fprintf(w, "gcs_emulator_requests_total{method=\"%s\",path=\"%s\"} %d\n", method, path, count)
			}
		}

		fmt.Fprintf(w, "# HELP gcs_emulator_request_duration_seconds Request duration in seconds\n")
		fmt.Fprintf(w, "# TYPE gcs_emulator_request_duration_seconds histogram\n")
		for _, d := range c.requestDuration {
			fmt.Fprintf(w, "gcs_emulator_request_duration_seconds %f\n", d)
		}

		objectsTotal := int64(0)
		if c.objectsTotalFunc != nil {
			objectsTotal = c.objectsTotalFunc()
		}
		fmt.Fprintf(w, "# HELP gcs_emulator_objects_total Total number of objects\n")
		fmt.Fprintf(w, "# TYPE gcs_emulator_objects_total gauge\n")
		fmt.Fprintf(w, "gcs_emulator_objects_total %d\n", objectsTotal)

		bucketsTotal := int64(0)
		if c.bucketsTotalFunc != nil {
			bucketsTotal = c.bucketsTotalFunc()
		}
		fmt.Fprintf(w, "# HELP gcs_emulator_buckets_total Total number of buckets\n")
		fmt.Fprintf(w, "# TYPE gcs_emulator_buckets_total gauge\n")
		fmt.Fprintf(w, "gcs_emulator_buckets_total %d\n", bucketsTotal)
	}
}
