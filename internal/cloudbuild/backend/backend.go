package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudbuild/model"
)

type CloudBuildBackend interface {
	CreateTrigger(ctx context.Context, trigger *model.BuildTrigger) (*model.BuildTrigger, error)
	GetTrigger(ctx context.Context, name string) (*model.BuildTrigger, error)
	ListTriggers(ctx context.Context, project string) ([]*model.BuildTrigger, error)
	DeleteTrigger(ctx context.Context, name string) error
	RunTrigger(ctx context.Context, name string) (*model.Build, error)
	Shutdown() error
}
