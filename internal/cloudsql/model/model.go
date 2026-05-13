package model

type Instance struct {
	Name            string `json:"name"`
	DatabaseVersion string `json:"databaseVersion,omitempty"`
	State           string `json:"state,omitempty"`
	Tier            string `json:"tier,omitempty"`
	Region          string `json:"region,omitempty"`
}
