package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/bigquery/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/bigquery/model"
)

type Handler struct {
	Backend backend.BigQueryBackend
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/bigquery/v2/")
	parts := strings.Split(path, "/")

	if len(parts) >= 2 && parts[0] == "projects" {
		project := parts[1]

		if len(parts) == 3 && parts[2] == "datasets" {
			switch r.Method {
			case http.MethodPost: h.createDataset(w, r, project)
			case http.MethodGet: h.listDatasets(w, r, project)
			default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		if len(parts) >= 4 && parts[2] == "datasets" {
			ds := parts[3]
			if len(parts) == 5 && parts[4] == "tables" {
				switch r.Method {
				case http.MethodPost: h.createTable(w, r, project, ds)
				case http.MethodGet: h.listTables(w, r, project, ds)
				default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
				}
				return
			}
			if len(parts) >= 6 && parts[4] == "tables" {
				tbl := parts[5]
				if strings.HasSuffix(tbl, ":insertAll") {
					h.insertRows(w, r, project, ds, strings.TrimSuffix(tbl, ":insertAll"))
					return
				}
				switch r.Method {
				case http.MethodGet: h.getTable(w, r, project, ds, tbl)
				case http.MethodDelete: h.deleteTable(w, r, project, ds, tbl)
				default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
				}
				return
			}
			switch r.Method {
			case http.MethodGet: h.getDataset(w, r, project, ds)
			case http.MethodDelete: h.deleteDataset(w, r, project, ds)
			default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}
		if len(parts) == 2 && parts[1] == "queries" {
			if r.Method == http.MethodPost { h.query(w, r, project) }
			return
		}
	}

	writeError(w, http.StatusBadRequest, "Invalid path")
}

func (h *Handler) createDataset(w http.ResponseWriter, r *http.Request, project string) {
	var ds model.Dataset
	json.NewDecoder(r.Body).Decode(&ds)
	if ds.DatasetReference == nil { ds.DatasetReference = &model.DatasetReference{} }
	if ds.DatasetReference.ProjectID == "" { ds.DatasetReference.ProjectID = project }
	result, _ := h.Backend.CreateDataset(r.Context(), project, &ds)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getDataset(w http.ResponseWriter, r *http.Request, project, ds string) {
	d, err := h.Backend.GetDataset(r.Context(), project, ds)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, d)
}
func (h *Handler) listDatasets(w http.ResponseWriter, r *http.Request, project string) {
	dss, _ := h.Backend.ListDatasets(r.Context(), project)
	writeJSON(w, http.StatusOK, map[string]interface{}{"datasets": dss})
}
func (h *Handler) deleteDataset(w http.ResponseWriter, r *http.Request, project, ds string) {
	h.Backend.DeleteDataset(r.Context(), project, ds)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) createTable(w http.ResponseWriter, r *http.Request, project, ds string) {
	var tbl model.Table
	json.NewDecoder(r.Body).Decode(&tbl)
	if tbl.TableReference == nil { tbl.TableReference = &model.TableReference{} }
	if tbl.TableReference.ProjectID == "" { tbl.TableReference.ProjectID = project }
	if tbl.TableReference.DatasetID == "" { tbl.TableReference.DatasetID = ds }
	result, _ := h.Backend.CreateTable(r.Context(), project, ds, &tbl)
	writeJSON(w, http.StatusOK, result)
}
func (h *Handler) getTable(w http.ResponseWriter, r *http.Request, project, ds, tbl string) {
	t, err := h.Backend.GetTable(r.Context(), project, ds, tbl)
	if err != nil { writeError(w, http.StatusNotFound, err.Error()); return }
	writeJSON(w, http.StatusOK, t)
}
func (h *Handler) listTables(w http.ResponseWriter, r *http.Request, project, ds string) {
	tbls, _ := h.Backend.ListTables(r.Context(), project, ds)
	writeJSON(w, http.StatusOK, map[string]interface{}{"tables": tbls})
}
func (h *Handler) deleteTable(w http.ResponseWriter, r *http.Request, project, ds, tbl string) {
	h.Backend.DeleteTable(r.Context(), project, ds, tbl)
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) insertRows(w http.ResponseWriter, r *http.Request, project, ds, tbl string) {
	var req model.InsertRequest
	json.NewDecoder(r.Body).Decode(&req)
	h.Backend.InsertRows(r.Context(), project, ds, tbl, req.Rows)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
func (h *Handler) query(w http.ResponseWriter, r *http.Request, project string) {
	var req model.QueryRequest
	json.NewDecoder(r.Body).Decode(&req)
	if req.DefaultDataset == nil { req.DefaultDataset = &model.DatasetReference{} }
	if req.DefaultDataset.ProjectID == "" { req.DefaultDataset.ProjectID = project }
	_ = project
	result, err := h.Backend.Query(r.Context(), req.Query)
	if err != nil { writeError(w, http.StatusBadRequest, err.Error()); return }
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{"code": status, "message": message},
	})
}
