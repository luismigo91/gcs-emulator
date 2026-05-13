package backend
import("context";"errors";"sync";m "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudnat/model")
var Err=errors.New("not found")
type MB struct{mu sync.RWMutex;nats map[string]*m.NATConfig}
func New()*MB{return &MB{nats:map[string]*m.NATConfig{}}}
func(b*MB)Create(c context.Context,n*m.NATConfig)(*m.NATConfig,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.nats[n.Name];e{return nil,errors.New("exists")};b.nats[n.Name]=n;return n,nil}
func(b*MB)Get(c context.Context,nm string)(*m.NATConfig,error){b.mu.RLock();defer b.mu.RUnlock();n,e:=b.nats[nm];if!e{return nil,Err};return n,nil}
func(b*MB)List(c context.Context)([]*m.NATConfig,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.NATConfig,0,len(b.nats));for _,v:=range b.nats{r=append(r,v)};return r,nil}
func(b*MB)Delete(c context.Context,nm string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.nats,nm);return nil}
func(b*MB)Shutdown()error{return nil}
