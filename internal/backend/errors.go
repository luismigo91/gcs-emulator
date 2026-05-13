package backend

import "errors"

var (
	ErrBucketNotFound        = errors.New("bucket not found")
	ErrObjectNotFound        = errors.New("object not found")
	ErrBucketNotEmpty        = errors.New("bucket is not empty")
	ErrGenerationNotFound    = errors.New("generation not found")
	ErrBucketAlreadyExists   = errors.New("bucket already exists")
	ErrPreconditionFailed    = errors.New("precondition failed")
	ErrComposeTooManySources = errors.New("too many source objects for compose (max 32)")
	ErrUploadSessionNotFound = errors.New("upload session not found")
	ErrNotificationNotFound  = errors.New("notification not found")
)
