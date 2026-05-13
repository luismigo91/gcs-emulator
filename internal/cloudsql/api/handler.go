package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudsql/backend"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudsql/model"
)

type Handler struct{ Backend *backend.MB }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "projects" { writeError(w, http.StatusBadRequest, "Invalid path"); return }
	pid := parts[1]

	if len(parts) >= 3 && parts[2] == "instances" {
		iname := ""
		if len(parts) >= 4 { iname = parts[3] }
		switch {
		case iname == "" && r.Method == http.MethodPost: h.create(w, r, pid)
		case iname == "" && r.Method == http.MethodGet: h.list(w, r)
		case iname != "" && r.Method == http.MethodGet: h.get(w, r, iname)
		case iname != "" && r.Method == http.MethodDelete: h.del(w, r, iname)
		case iname != "" && r.Method == http.MethodPatch: h.update(w, r, iname)
		default: writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	writeError(w, http.StatusBadRequest, "Invalid path")
}

func(h*Handler)create(w http.ResponseWriter,r *http.Request,pid string){var i model.Instance;json.NewDecoder(r.Body).Decode(&i);if i.Name==""{i.Name=r.URL.Query().Get("instanceId")};result,_:=h.Backend.Create(r.Context(),&i);json.NewEncoder(w).Encode(result)}
func(h*Handler)get(w http.ResponseWriter,r *http.Request,n string){i,err:=h.Backend.Get(r.Context(),n);if err!=nil{writeError(w,404,err.Error());return};w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(i)}
func(h*Handler)list(w http.ResponseWriter,r *http.Request){is,_:=h.Backend.List(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"instances":is})}
func(h*Handler)del(w http.ResponseWriter,r *http.Request,n string){h.Backend.Delete(r.Context(),n);w.WriteHeader(200)}
func(h*Handler)update(w http.ResponseWriter,r *http.Request,n string){var u model.Instance;json.NewDecoder(r.Body).Decode(&u);h.Backend.Create(r.Context(),&u);w.WriteHeader(200)}

func writeError(w http.ResponseWriter,s int,m string){w.Header().Set("Content-Type","application/json");w.WriteHeader(s);json.NewEncoder(w).Encode(map[string]interface{}{"error":map[string]interface{}{"code":s,"message":m}})}
