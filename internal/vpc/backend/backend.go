package backend

import("context";"errors";"sync";m "github.com/luismiguelgilolivert/gcs-emulator/internal/vpc/model")
var Err=errors.New("not found")
type MB struct{mu sync.RWMutex;nets map[string]*m.Network;subs map[string]*m.Subnetwork}
func New()*MB{return &MB{nets:map[string]*m.Network{},subs:map[string]*m.Subnetwork{}}}
func(b*MB)CreateNetwork(c context.Context,n*m.Network)(*m.Network,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.nets[n.Name];e{return nil,errors.New("exists")};b.nets[n.Name]=n;return n,nil}
func(b*MB)GetNetwork(c context.Context,nm string)(*m.Network,error){b.mu.RLock();defer b.mu.RUnlock();n,e:=b.nets[nm];if!e{return nil,Err};return n,nil}
func(b*MB)ListNetworks(c context.Context)([]*m.Network,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.Network,0,len(b.nets));for _,n:=range b.nets{r=append(r,n)};return r,nil}
func(b*MB)DeleteNetwork(c context.Context,nm string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.nets,nm);return nil}
func(b*MB)CreateSubnet(c context.Context,s*m.Subnetwork)(*m.Subnetwork,error){b.mu.Lock();defer b.mu.Unlock();if _,e:=b.subs[s.Name];e{return nil,errors.New("exists")};b.subs[s.Name]=s;return s,nil}
func(b*MB)ListSubnets(c context.Context)([]*m.Subnetwork,error){b.mu.RLock();defer b.mu.RUnlock();r:=make([]*m.Subnetwork,0,len(b.subs));for _,s:=range b.subs{r=append(r,s)};return r,nil}
func(b*MB)DeleteSubnet(c context.Context,nm string)error{b.mu.Lock();defer b.mu.Unlock();delete(b.subs,nm);return nil}
func(b*MB)Shutdown()error{return nil}
