package model
type Router struct{Name string`json:"name"`;Network string`json:"network"`;Region string`json:"region"`}
type NAT struct{Name string`json:"name"`;Router string`json:"router"`;Region string`json:"region,omitempty"`}
