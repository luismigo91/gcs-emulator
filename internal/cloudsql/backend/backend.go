package backend

import (
	"context"
	"errors"
	"sync"
	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudsql/model"
)

var Err = errors.New("not found")

type MB struct{ mu sync.RWMutex; inst map[string]*model.Instance }
func New()*MB{ return &MB{inst:map[string]*model.Instance{}} }
func(b*MB)Create(ctx context.Context,i*model.Instance)(*model.Instance,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.inst[i.Name];e{return nil,errors.New("exists")};i.State="RUNNABLE";b.inst[i.Name]=i;return i,nil}
func(b*MB)Get(ctx context.Context,n string)(*model.Instance,error){b.mu.RLock();defer b.mu.RUnlock();i,e:=b.inst[n];if!e{return nil,Err};return i,nil}
func(b*MB)List(ctx context.Context)([]*model.Instance,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*model.Instance,0,len(b.inst));for _,i:=range b.inst{r=append(r,i)};return r,nil}
func(b*MB)Delete(ctx context.Context,n string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.inst,n);return nil}
func(b*MB)Shutdown()error{return nil}
