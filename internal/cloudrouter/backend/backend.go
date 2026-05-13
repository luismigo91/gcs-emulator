package backend
import("context";"errors";"sync";m "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudrouter/model")
var Err=errors.New("not found")
type MB struct{mu sync.RWMutex;routers map[string]*m.Router;nats map[string]*m.NAT}
func New()*MB{return &MB{routers:map[string]*m.Router{},nats:map[string]*m.NAT{}}}
func(b*MB)CreateRouter(c context.Context,r*m.Router)(*m.Router,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.routers[r.Name];e{return nil,errors.New("exists")};b.routers[r.Name]=r;return r,nil}
func(b*MB)GetRouter(c context.Context,n string)(*m.Router,error){b.mu.RLock();defer b.mu.RUnlock();r,e:=b.routers[n];if!e{return nil,Err};return r,nil}
func(b*MB)ListRouters(c context.Context)([]*m.Router,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.Router,0,len(b.routers));for _,v:=range b.routers{r=append(r,v)};return r,nil}
func(b*MB)DeleteRouter(c context.Context,n string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.routers,n);return nil}
func(b*MB)CreateNAT(c context.Context,n*m.NAT)(*m.NAT,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.nats[n.Name];e{return nil,errors.New("exists")};b.nats[n.Name]=n;return n,nil}
func(b*MB)ListNATs(c context.Context)([]*m.NAT,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.NAT,0,len(b.nats));for _,v:=range b.nats{r=append(r,v)};return r,nil}
func(b*MB)DeleteNAT(c context.Context,n string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.nats,n);return nil}
func(b*MB)Shutdown()error{return nil}
