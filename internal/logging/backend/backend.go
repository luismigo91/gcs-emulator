package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/logging/model"
)

type LoggingBackend interface {
	Write(ctx context.Context, entries []*model.LogEntry) error
	List(ctx context.Context, logName string, pageSize int) ([]*model.LogEntry, error)
	Shutdown() error
}
