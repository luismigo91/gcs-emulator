package model

import "time"

type Certificate struct {
	Name        string            `json:"name"`
	Domains     []string          `json:"domains"`
	Scope       string            `json:"scope,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Description string            `json:"description,omitempty"`
	CreateTime  time.Time         `json:"createTime,omitempty"`
	ExpireTime  time.Time         `json:"expireTime,omitempty"`
}

type DnsAuthorization struct {
	Name       string `json:"name"`
	Domain     string `json:"domain"`
	DnsRecord  *DnsRecord `json:"dnsRecord,omitempty"`
}

type DnsRecord struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
}
