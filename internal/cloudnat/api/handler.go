package api
import("encoding/json";"net/http";b "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudnat/backend";m "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudnat/model")
type H struct{BE *b.MB}
func(h*H)ServeHTTP(w http.ResponseWriter,r *http.Request){
	switch r.Method {
	case "POST":
		var n m.NATConfig;json.NewDecoder(r.Body).Decode(&n);if n.Name==""{n.Name=r.URL.Query().Get("natId")};h.BE.Create(r.Context(),&n)
		w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(n)
	case "GET":
		ns,_:=h.BE.List(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"nats":ns})
	case "DELETE":
		h.BE.Delete(r.Context(),"default");w.WriteHeader(200)
	default:w.WriteHeader(405)
	}
}
