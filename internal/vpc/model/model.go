package model

type Network struct {
	Name                  string `json:"name"`
	AutoCreateSubnetworks bool   `json:"autoCreateSubnetworks"`
	Subnetworks           []string `json:"subnetworks,omitempty"`
}

type Subnetwork struct {
	Name        string `json:"name"`
	Network     string `json:"network"`
	IPCidrRange string `json:"ipCidrRange"`
	Region      string `json:"region"`
}
