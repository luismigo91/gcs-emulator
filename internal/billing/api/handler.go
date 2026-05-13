package api

import (
	"encoding/json"
	"net/http"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/billing/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/billing/model"
)

type Handler struct{ Backend *backend.MemoryBillingBackend }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var b model.Budget
		json.NewDecoder(r.Body).Decode(&b)
		if b.Name == "" { b.Name = r.URL.Query().Get("budgetId") }
		result, _ := h.Backend.CreateBudget(r.Context(), &b)
		writeJSON(w, http.StatusOK, result)
	case http.MethodGet:
		bs, _ := h.Backend.ListBudgets(r.Context())
		writeJSON(w, http.StatusOK, map[string]interface{}{"budgets": bs})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
func writeJSON(w http.ResponseWriter, s int, d interface{}) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(d) }
func writeError(w http.ResponseWriter, s int, m string) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(s); json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]interface{}{"code": s, "message": m}}) }
