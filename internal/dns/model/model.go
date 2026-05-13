package model

type ManagedZone struct {
	Name        string   `json:"name"`
	DNSName     string   `json:"dnsName"`
	Description string   `json:"description,omitempty"`
	NameServers []string `json:"nameServers,omitempty"`
}

type ResourceRecordSet struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	TTL     int      `json:"ttl"`
	Rrdatas []string `json:"rrdatas"`
}

type Change struct {
	Additions []*ResourceRecordSet `json:"additions,omitempty"`
	Deletions []*ResourceRecordSet `json:"deletions,omitempty"`
}
