package backend
import("context";"errors";"sync";m "github.com/luismiguelgilolivert/gcs-emulator/internal/cloudarmor/model")
var Err=errors.New("not found")
type MB struct{mu sync.RWMutex;pol map[string]*m.Policy}
func New()*MB{return &MB{pol:map[string]*m.Policy{}}}
func(b*MB)Create(c context.Context,p*m.Policy)(*m.Policy,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.pol[p.Name];e{return nil,errors.New("exists")};b.pol[p.Name]=p;return p,nil}
func(b*MB)Get(c context.Context,n string)(*m.Policy,error){b.mu.RLock();defer b.mu.RUnlock();p,e:=b.pol[n];if!e{return nil,Err};return p,nil}
func(b*MB)List(c context.Context)([]*m.Policy,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.Policy,0,len(b.pol));for _,p:=range b.pol{r=append(r,p)};return r,nil}
func(b*MB)Delete(c context.Context,n string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.pol,n);return nil}
func(b*MB)Shutdown()error{return nil}
