package model

type Instance struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName,omitempty"`
	MemorySizeGb int    `json:"memorySizeGb"`
	Tier         string `json:"tier,omitempty"`
	State        string `json:"state,omitempty"`
}
