package model

import "time"

type KeyRing struct {
	Name       string    `json:"name"`
	CreateTime time.Time `json:"createTime,omitempty"`
}

type CryptoKey struct {
	Name            string            `json:"name"`
	Purpose         string            `json:"purpose"`
	CreateTime      time.Time         `json:"createTime,omitempty"`
	NextRotationTime *string          `json:"nextRotationTime,omitempty"`
	RotationPeriod  *string           `json:"rotationPeriod,omitempty"`
	VersionTemplate *VersionTemplate  `json:"versionTemplate,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

type VersionTemplate struct {
	Algorithm       string `json:"algorithm"`
	ProtectionLevel string `json:"protectionLevel"`
}

type CryptoKeyVersion struct {
	Name         string    `json:"name"`
	State        string    `json:"state"`
	Algorithm    string    `json:"algorithm"`
	CreateTime   time.Time `json:"createTime,omitempty"`
	DestroyTime  time.Time `json:"destroyTime,omitempty"`
	DestroyEventTime time.Time `json:"destroyEventTime,omitempty"`
}

type EncryptRequest struct {
	Plaintext []byte `json:"plaintext"`
}

type EncryptResponse struct {
	Ciphertext []byte `json:"ciphertext"`
}

type DecryptRequest struct {
	Ciphertext []byte `json:"ciphertext"`
}

type DecryptResponse struct {
	Plaintext []byte `json:"plaintext"`
}
