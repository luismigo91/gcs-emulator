package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/secretmanager/model"
)

type SecretManagerBackend interface {
	CreateSecret(ctx context.Context, secret *model.Secret) (*model.Secret, error)
	GetSecret(ctx context.Context, project, name string) (*model.Secret, error)
	ListSecrets(ctx context.Context, project string) ([]*model.Secret, error)
	UpdateSecret(ctx context.Context, project, name string, secret *model.Secret, paths []string) (*model.Secret, error)
	DeleteSecret(ctx context.Context, project, name string) error

	AddVersion(ctx context.Context, project, name string, payload []byte) (*model.SecretVersion, error)
	ListVersions(ctx context.Context, project, name string) ([]*model.SecretVersion, error)
	GetVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error)
	AccessVersion(ctx context.Context, project, name, version string) ([]byte, error)
	EnableVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error)
	DisableVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error)
	DestroyVersion(ctx context.Context, project, name, version string) (*model.SecretVersion, error)

	Shutdown() error
}
