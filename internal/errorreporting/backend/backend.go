package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/errorreporting/model"
)

type ErrorReportingBackend interface {
	Report(ctx context.Context, event *model.ReportedErrorEvent) error
	List(ctx context.Context, pageSize int) ([]*model.ReportedErrorEvent, error)
	Shutdown() error
}
