package model

import "time"

type ServiceAccount struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

type SignJwtRequest struct {
	Payload string `json:"payload"`
}

type SignJwtResponse struct {
	KeyID     string `json:"keyId"`
	SignedJwt string `json:"signedJwt"`
}

type GenerateAccessTokenResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpireTime  time.Time `json:"expireTime"`
}

type CreateKeyRequest struct {
	PrivateKeyType string `json:"privateKeyType"`
	KeyAlgorithm   string `json:"keyAlgorithm"`
}

type ServiceAccountKey struct {
	Name        string `json:"name"`
	PrivateKey  string `json:"privateKeyData"`
	KeyType     string `json:"privateKeyType"`
	KeyAlgorithm string `json:"keyAlgorithm"`
	ValidAfter  time.Time `json:"validAfterTime,omitempty"`
}
