package admin

import (
	"encoding/json"
	"net/http"
)

func ServicesHandler(services []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"emulator": "gcs-emulator",
			"services": services,
		})
	}
}
