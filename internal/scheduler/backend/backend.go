package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/scheduler/model"
)

type SchedulerBackend interface {
	CreateJob(ctx context.Context, job *model.Job) (*model.Job, error)
	GetJob(ctx context.Context, name string) (*model.Job, error)
	ListJobs(ctx context.Context, project, location string) ([]*model.Job, error)
	UpdateJob(ctx context.Context, name string, job *model.Job) (*model.Job, error)
	DeleteJob(ctx context.Context, name string) error
	PauseJob(ctx context.Context, name string) (*model.Job, error)
	ResumeJob(ctx context.Context, name string) (*model.Job, error)
	RunJob(ctx context.Context, name string) error
	Shutdown() error
}
