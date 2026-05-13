package model

import "time"

type BucketAttrs struct {
	Name             string            `json:"name"`
	ProjectNumber    string            `json:"projectNumber,omitempty"`
	Location         string            `json:"location,omitempty"`
	StorageClass     string            `json:"storageClass,omitempty"`
	Versioning       *Versioning       `json:"versioning,omitempty"`
	Lifecycle        *Lifecycle        `json:"lifecycle,omitempty"`
	IamConfiguration *IamConfiguration `json:"iamConfiguration,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
}

type ObjectAttrs struct {
	Name         string            `json:"name"`
	ContentType  string            `json:"contentType,omitempty"`
	StorageClass string            `json:"storageClass,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type ObjectUpdateAttrs struct {
	ContentType        *string           `json:"contentType,omitempty"`
	ContentEncoding    *string           `json:"contentEncoding,omitempty"`
	ContentDisposition *string           `json:"contentDisposition,omitempty"`
	ContentLanguage    *string           `json:"contentLanguage,omitempty"`
	CacheControl       *string           `json:"cacheControl,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	EventBasedHold     *bool             `json:"eventBasedHold,omitempty"`
	TemporaryHold      *bool             `json:"temporaryHold,omitempty"`
}

type BucketUpdateAttrs struct {
	Versioning       *Versioning       `json:"versioning,omitempty"`
	Lifecycle        *Lifecycle        `json:"lifecycle,omitempty"`
	IamConfiguration *IamConfiguration `json:"iamConfiguration,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	RetentionPolicy  *RetentionPolicy  `json:"retentionPolicy,omitempty"`
	CORS             []CORSRule        `json:"cors,omitempty"`
}

type Project struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	TimeCreated  time.Time `json:"timeCreated"`
}
