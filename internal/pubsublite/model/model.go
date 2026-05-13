package model

type Reservation struct {
	Name               string `json:"name"`
	ThroughputCapacity int64  `json:"throughputCapacity,string,omitempty"`
}

type LiteTopic struct {
	Name       string `json:"name"`
	Partitions int64  `json:"partitions,string,omitempty"`
}
