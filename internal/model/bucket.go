package model

import "time"

type Versioning struct {
	Enabled bool `json:"enabled"`
}

type LifecycleRuleAction struct {
	Type         string `json:"type"`
	StorageClass string `json:"storageClass,omitempty"`
}

type LifecycleRuleCondition struct {
	Age               int       `json:"age,omitempty"`
	CreatedBefore     string    `json:"createdBefore,omitempty"`
	IsLive            *bool     `json:"isLive,omitempty"`
	NumNewerVersions  int       `json:"numNewerVersions,omitempty"`
	MatchesStorageClass []string `json:"matchesStorageClass,omitempty"`
}

type LifecycleRule struct {
	Action    LifecycleRuleAction    `json:"action"`
	Condition LifecycleRuleCondition `json:"condition"`
}

type Lifecycle struct {
	Rule []LifecycleRule `json:"rule"`
}

type IamConfiguration struct {
	UniformBucketLevelAccess UniformBucketLevelAccess `json:"uniformBucketLevelAccess"`
	PublicAccessPrevention   string                   `json:"publicAccessPrevention,omitempty"`
}

type UniformBucketLevelAccess struct {
	Enabled    bool   `json:"enabled"`
	LockedTime string `json:"lockedTime,omitempty"`
}

type Bucket struct {
	Kind                     string           `json:"kind"`
	ID                       string           `json:"id"`
	SelfLink                 string           `json:"selfLink,omitempty"`
	ProjectNumber            uint64           `json:"projectNumber,omitempty,string"`
	ProjectID                string           `json:"-"`
	Name                     string           `json:"name"`
	TimeCreated              time.Time        `json:"timeCreated"`
	Updated                  time.Time        `json:"updated"`
	Metageneration           int64            `json:"metageneration,string"`
	Location                 string           `json:"location,omitempty"`
	LocationType             string           `json:"locationType,omitempty"`
	StorageClass             string           `json:"storageClass,omitempty"`
	Versioning               *Versioning      `json:"versioning,omitempty"`
	Lifecycle                *Lifecycle       `json:"lifecycle,omitempty"`
	IamConfiguration         *IamConfiguration `json:"iamConfiguration,omitempty"`
	Labels                   map[string]string `json:"labels,omitempty"`
	RetentionPolicy          *RetentionPolicy  `json:"retentionPolicy,omitempty"`
	CORS                     []CORSRule       `json:"cors,omitempty"`
	Website                  *WebsiteConfig   `json:"website,omitempty"`
	Encryption               *EncryptionConfig `json:"encryption,omitempty"`
}

type CORSRule struct {
	Origin         []string `json:"origin,omitempty"`
	Method         []string `json:"method,omitempty"`
	ResponseHeader []string `json:"responseHeader,omitempty"`
	MaxAgeSeconds  int      `json:"maxAgeSeconds,omitempty"`
}

type RetentionPolicy struct {
	RetentionPeriod string    `json:"retentionPeriod,omitempty"`
	EffectiveTime   time.Time `json:"effectiveTime,omitempty"`
	IsLocked        bool      `json:"isLocked,omitempty"`
}

type BucketACL struct {
	Kind     string `json:"kind,omitempty"`
	ID       string `json:"id,omitempty"`
	SelfLink string `json:"selfLink,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Entity   string `json:"entity"`
	Role     string `json:"role"`
	Email    string `json:"email,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Project  *ProjectACL `json:"projectTeam,omitempty"`
	ETag     string `json:"etag,omitempty"`
}

type ProjectACL struct {
	ProjectNumber string `json:"projectNumber,omitempty"`
	Team          string `json:"team,omitempty"`
}

type ObjectACL struct {
	Kind     string `json:"kind,omitempty"`
	ID       string `json:"id,omitempty"`
	SelfLink string `json:"selfLink,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Object   string `json:"object,omitempty"`
	Generation int64 `json:"generation,omitempty,string"`
	Entity   string `json:"entity"`
	Role     string `json:"role"`
	Email    string `json:"email,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Project  *ProjectACL `json:"projectTeam,omitempty"`
	ETag     string `json:"etag,omitempty"`
}

type WebsiteConfig struct {
	MainPageSuffix string `json:"mainPageSuffix,omitempty"`
	NotFoundPage   string `json:"notFoundPage,omitempty"`
}

type EncryptionConfig struct {
	DefaultKMSKeyName string `json:"defaultKmsKeyName,omitempty"`
}
