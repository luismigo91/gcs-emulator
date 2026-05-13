package api

import("encoding/json";"net/http";"strings";b "github.com/luismiguelgilolivert/gcs-emulator/internal/vpc/backend";m "github.com/luismiguelgilolivert/gcs-emulator/internal/vpc/model")
type H struct{BE *b.MB}

func(h*H)ServeHTTP(w http.ResponseWriter,r *http.Request){
	p:=strings.TrimPrefix(r.URL.Path,"/compute/v1/")
	parts:=strings.Split(p,"/")
	if len(parts)<2||parts[0]!="projects"{w.WriteHeader(400);return}
	pid:=parts[1]
	_=pid

	isNet:=len(parts)>=4&&parts[2]=="global"&&parts[3]=="networks"
	isSub:=len(parts)>=5&&parts[2]=="regions"&&parts[4]=="subnetworks"

	if isNet {
		nname:="";if len(parts)>=5{nname=parts[4]}
		switch{
		case nname==""&&r.Method=="POST":h.createNet(w,r)
		case nname==""&&r.Method=="GET":h.listNets(w,r)
		case nname!=""&&r.Method=="GET":h.getNet(w,r,nname)
		case nname!=""&&r.Method=="DELETE":h.delNet(w,r,nname)
		default:w.WriteHeader(405)
		}
		return
	}
	if isSub {
		sname:="";if len(parts)>=6{sname=parts[5]}
		switch{
		case sname==""&&r.Method=="POST":h.createSub(w,r)
		case sname==""&&r.Method=="GET":h.listSubs(w,r)
		case sname!=""&&r.Method=="DELETE":h.delSub(w,r,sname)
		default:w.WriteHeader(405)
		}
		return
	}
	w.WriteHeader(400)
}

func(h*H)createNet(w http.ResponseWriter,r *http.Request){var n m.Network;json.NewDecoder(r.Body).Decode(&n);if n.Name==""{n.Name=r.URL.Query().Get("networkId")};h.BE.CreateNetwork(r.Context(),&n);w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(n)}
func(h*H)getNet(w http.ResponseWriter,r *http.Request,nm string){n,_:=h.BE.GetNetwork(r.Context(),nm);w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(n)}
func(h*H)listNets(w http.ResponseWriter,r *http.Request){ns,_:=h.BE.ListNetworks(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"items":ns})}
func(h*H)delNet(w http.ResponseWriter,r *http.Request,nm string){h.BE.DeleteNetwork(r.Context(),nm);w.WriteHeader(200)}
func(h*H)createSub(w http.ResponseWriter,r *http.Request){var s m.Subnetwork;json.NewDecoder(r.Body).Decode(&s);h.BE.CreateSubnet(r.Context(),&s);w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(s)}
func(h*H)listSubs(w http.ResponseWriter,r *http.Request){ss,_:=h.BE.ListSubnets(r.Context());w.Header().Set("Content-Type","application/json");w.WriteHeader(200);json.NewEncoder(w).Encode(map[string]interface{}{"items":ss})}
func(h*H)delSub(w http.ResponseWriter,r *http.Request,nm string){h.BE.DeleteSubnet(r.Context(),nm);w.WriteHeader(200)}
