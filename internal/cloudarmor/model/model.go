package model
type Policy struct{Name string`json:"name"`;Rules []Rule`json:"rules,omitempty"`}
type Rule struct{Action string`json:"action"`;Priority int`json:"priority"`;Description string`json:"description,omitempty"`}
