package model

import "time"

type Owner struct {
	Entity string `json:"entity,omitempty"`
	EntityID string `json:"entityId,omitempty"`
}

type CustomerEncryption struct {
	EncryptionAlgorithm string `json:"encryptionAlgorithm"`
	KeySha256           string `json:"keySha256"`
}

type Object struct {
	Kind                    string               `json:"kind"`
	ID                      string               `json:"id"`
	SelfLink                string               `json:"selfLink,omitempty"`
	MediaLink               string               `json:"mediaLink,omitempty"`
	Name                    string               `json:"name"`
	Bucket                  string               `json:"bucket"`
	Generation              int64                `json:"generation,omitempty,string"`
	Metageneration          int64                `json:"metageneration,string"`
	ContentType             string               `json:"contentType,omitempty"`
	ContentEncoding         string               `json:"contentEncoding,omitempty"`
	ContentDisposition      string               `json:"contentDisposition,omitempty"`
	ContentLanguage         string               `json:"contentLanguage,omitempty"`
	CacheControl            string               `json:"cacheControl,omitempty"`
	StorageClass            string               `json:"storageClass,omitempty"`
	Size                    int64                `json:"size,string,omitempty"`
	MD5Hash                 string               `json:"md5Hash,omitempty"`
	CRC32C                  string               `json:"crc32c,omitempty"`
	ETag                    string               `json:"etag,omitempty"`
	TimeCreated             time.Time            `json:"timeCreated"`
	Updated                 time.Time            `json:"updated"`
	TimeDeleted             time.Time            `json:"timeDeleted,omitempty"`
	TimeStorageClassUpdated time.Time            `json:"timeStorageClassUpdated,omitempty"`
	CustomTime              time.Time            `json:"customTime,omitempty"`
	Metadata                map[string]string    `json:"metadata,omitempty"`
	Owner                   *Owner               `json:"owner,omitempty"`
	CustomerEncryption      *CustomerEncryption  `json:"customerEncryption,omitempty"`
	EventBasedHold          bool                 `json:"eventBasedHold,omitempty"`
	TemporaryHold           bool                 `json:"temporaryHold,omitempty"`
	RetentionExpirationTime time.Time            `json:"retentionExpirationTime,omitempty"`
	ComponentCount          int                  `json:"componentCount,omitempty"`
}
