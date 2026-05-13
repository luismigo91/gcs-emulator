package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/monitoring/model"
)

type MonitoringBackend interface {
	CreateTimeSeries(ctx context.Context, series []*model.TimeSeries) error
	ListTimeSeries(ctx context.Context, project, filter string, pageSize int) ([]*model.TimeSeries, error)
	Shutdown() error
}
