package backend

import ("context";"errors";"sync";m "github.com/luismiguelgilolivert/gcs-emulator/internal/memorystore/model")
var Err=errors.New("not found")
type MB struct{mu sync.RWMutex;inst map[string]*m.Instance}
func New()*MB{return &MB{inst:map[string]*m.Instance{}}}
func(b*MB)Create(c context.Context,i*m.Instance)(*m.Instance,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.inst[i.Name];e{return nil,errors.New("exists")};i.State="READY";b.inst[i.Name]=i;return i,nil}
func(b*MB)Get(c context.Context,n string)(*m.Instance,error){b.mu.RLock();defer b.mu.RUnlock();i,e:=b.inst[n];if!e{return nil,Err};return i,nil}
func(b*MB)List(c context.Context)([]*m.Instance,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.Instance,0,len(b.inst));for _,i:=range b.inst{r=append(r,i)};return r,nil}
func(b*MB)Delete(c context.Context,n string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.inst,n);return nil}
func(b*MB)Shutdown()error{return nil}
