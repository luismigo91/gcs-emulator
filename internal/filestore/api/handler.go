package api
import("encoding/json";"net/http";b "github.com/luismiguelgilolivert/gcs-emulator/internal/filestore/backend";m "github.com/luismiguelgilolivert/gcs-emulator/internal/filestore/model")
type H struct{BE *b.MB}
func(h*H)ServeHTTP(w http.ResponseWriter,r *http.Request){
	switch r.Method {
	case "POST":
		var i m.Instance;json.NewDecoder(r.Body).Decode(&i);if i.Name==""{i.Name=r.URL.Query().Get("instanceId")};h.BE.Create(r.Context(),&i)
		w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(i)
	case "GET":
		is,_:=h.BE.List(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"instances":is})
	case "DELETE":
		h.BE.Delete(r.Context(),"default");w.WriteHeader(200)
	default:w.WriteHeader(405)
	}
}
