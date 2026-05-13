package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/cloudtasks/model"
)

type CloudTasksBackend interface {
	CreateQueue(ctx context.Context, queue *model.Queue) (*model.Queue, error)
	GetQueue(ctx context.Context, name string) (*model.Queue, error)
	ListQueues(ctx context.Context, project, location string) ([]*model.Queue, error)
	UpdateQueue(ctx context.Context, name string, queue *model.Queue) (*model.Queue, error)
	DeleteQueue(ctx context.Context, name string) error
	PauseQueue(ctx context.Context, name string) (*model.Queue, error)
	ResumeQueue(ctx context.Context, name string) (*model.Queue, error)
	PurgeQueue(ctx context.Context, name string) error

	CreateTask(ctx context.Context, parent string, task *model.Task) (*model.Task, error)
	GetTask(ctx context.Context, name string) (*model.Task, error)
	ListTasks(ctx context.Context, parent string) ([]*model.Task, error)
	DeleteTask(ctx context.Context, name string) error
	RunTask(ctx context.Context, name string) (*model.Task, error)

	Shutdown() error
}
