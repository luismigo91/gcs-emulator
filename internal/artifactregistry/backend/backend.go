package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/artifactregistry/model"
)

type ArtifactRegistryBackend interface {
	CreateRepository(ctx context.Context, repo *model.Repository) (*model.Repository, error)
	GetRepository(ctx context.Context, name string) (*model.Repository, error)
	ListRepositories(ctx context.Context, parent string) ([]*model.Repository, error)
	DeleteRepository(ctx context.Context, name string) error
	ListDockerImages(ctx context.Context, parent string) ([]*model.DockerImage, error)
	Shutdown() error
}
