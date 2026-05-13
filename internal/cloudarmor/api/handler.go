package api
import("encoding/json";"net/http";b "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudarmor/backend";m "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudarmor/model")
type H struct{BE *b.MB}
func(h*H)ServeHTTP(w http.ResponseWriter,r *http.Request){
	switch r.Method {
	case "POST":
		var p m.Policy;json.NewDecoder(r.Body).Decode(&p);if p.Name==""{p.Name=r.URL.Query().Get("securityPolicyId")};h.BE.Create(r.Context(),&p)
		w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(p)
	case "GET":
		pols,_:=h.BE.List(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"items":pols})
	case "DELETE":
		h.BE.Delete(r.Context(),"default");w.WriteHeader(200)
	default:w.WriteHeader(405)
	}
}
