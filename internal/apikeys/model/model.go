package model

type Key struct {
	Name       string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	KeyString  string `json:"keyString,omitempty"`
	Restrictions *Restrictions `json:"restrictions,omitempty"`
}

type Restrictions struct {
	API_targets []string `json:"apiTargets,omitempty"`
}
