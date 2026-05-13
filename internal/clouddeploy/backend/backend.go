package backend
import("context";"errors";"sync";m "github.com/luismiguelgilolivert/gcs-emulator/internal/clouddeploy/model")
var Err=errors.New("not found")
type MB struct{mu sync.RWMutex;pipe map[string]*m.Pipeline}
func New()*MB{return &MB{pipe:map[string]*m.Pipeline{}}}
func(b*MB)Create(c context.Context,p*m.Pipeline)(*m.Pipeline,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.pipe[p.Name];e{return nil,errors.New("exists")};p.State="ACTIVE";b.pipe[p.Name]=p;return p,nil}
func(b*MB)Get(c context.Context,n string)(*m.Pipeline,error){b.mu.RLock();defer b.mu.RUnlock();p,e:=b.pipe[n];if!e{return nil,Err};return p,nil}
func(b*MB)List(c context.Context)([]*m.Pipeline,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.Pipeline,0,len(b.pipe));for _,p:=range b.pipe{r=append(r,p)};return r,nil}
func(b*MB)Delete(c context.Context,n string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.pipe,n);return nil}
func(b*MB)Shutdown()error{return nil}
