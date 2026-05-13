package api
import("encoding/json";"net/http";"strings";b "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudrouter/backend";m "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudrouter/model")
type H struct{BE *b.MB}
func(h*H)ServeHTTP(w http.ResponseWriter,r *http.Request){
	p:=r.URL.Path
	if strings.Contains(p,"/routers")&&!strings.Contains(p,"/nats"){
		switch r.Method{
		case "POST":var rt m.Router;json.NewDecoder(r.Body).Decode(&rt);if rt.Name==""{rt.Name=r.URL.Query().Get("routerId")};h.BE.CreateRouter(r.Context(),&rt);w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(rt)
		case "GET":rts,_:=h.BE.ListRouters(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"routers":rts})
		case "DELETE":h.BE.DeleteRouter(r.Context(),"default");w.WriteHeader(200)
		default:w.WriteHeader(405)
		}
		return
	}
	if strings.Contains(p,"/nat")||strings.Contains(p,"/nats"){
		switch r.Method{
		case "POST":var n m.NAT;json.NewDecoder(r.Body).Decode(&n);if n.Name==""{n.Name=r.URL.Query().Get("natId")};h.BE.CreateNAT(r.Context(),&n);w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(n)
		case "GET":ns,_:=h.BE.ListNATs(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"nats":ns})
		case "DELETE":h.BE.DeleteNAT(r.Context(),"default");w.WriteHeader(200)
		default:w.WriteHeader(405)
		}
		return
	}
	w.WriteHeader(400)
}
