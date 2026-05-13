package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/trace/model"
)

type TraceBackend interface {
	BatchWrite(ctx context.Context, spans []*model.Span) error
	List(ctx context.Context, pageSize int) ([]*model.Span, error)
	GetTrace(ctx context.Context, traceID string) ([]*model.Span, error)
	Shutdown() error
}
