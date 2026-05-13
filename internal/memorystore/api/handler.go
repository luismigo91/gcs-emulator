package api

import("encoding/json";"net/http";"strings";b "github.com/luismiguelgilolivert/gcs-emulator/internal/memorystore/backend";m "github.com/luismiguelgilolivert/gcs-emulator/internal/memorystore/model")
type H struct{BE *b.MB}

func(h*H)ServeHTTP(w http.ResponseWriter,r *http.Request){
	p:=r.URL.Path
	switch r.Method {
	case "POST":
		var i m.Instance;json.NewDecoder(r.Body).Decode(&i)
		if i.Name==""{i.Name=r.URL.Query().Get("instanceId")}
		r,_:=h.BE.Create(r.Context(),&i)
		w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(r)
	case "GET":
		if p!="/v1/projects/-/locations/-/instances" {
			parts:=strings.Split(strings.TrimPrefix(p,"/v1/"),"/")
			if len(parts)>=4 {
				i,err:=h.BE.Get(r.Context(),parts[3])
				if err!=nil{w.Header().Set("Content-Type","application/json");w.WriteHeader(404);json.NewEncoder(w).Encode(map[string]string{"error":err.Error()});return}
				w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(i);return
			}
		}
		is,_:=h.BE.List(r.Context())
		w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"instances":is})
	case "DELETE":
		parts:=strings.Split(strings.TrimPrefix(p,"/v1/"),"/")
		if len(parts)>=4{h.BE.Delete(r.Context(),parts[3])};w.WriteHeader(200)
	}
}
