package model

import "time"

type Secret struct {
	Name       string            `json:"name"`
	Replication *Replication     `json:"replication,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	CreateTime time.Time         `json:"createTime,omitempty"`
}

type Replication struct {
	Automatic *Automatic `json:"automatic,omitempty"`
}

type Automatic struct{}

type SecretVersion struct {
	Name        string    `json:"name"`
	CreateTime  time.Time `json:"createTime,omitempty"`
	State       string    `json:"state"`
	DestroyTime time.Time `json:"destroyTime,omitempty"`
}

type AccessResponse struct {
	Name    string    `json:"name"`
	Payload *Payload  `json:"payload"`
}

type Payload struct {
	Data []byte `json:"data"`
}

type AddVersionRequest struct {
	Payload *Payload `json:"payload"`
}

type UpdateSecretRequest struct {
	Secret *Secret  `json:"secret"`
	UpdateMask *UpdateMask `json:"updateMask"`
}

type UpdateMask struct {
	Paths []string `json:"paths"`
}
