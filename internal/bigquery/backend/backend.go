package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/bigquery/model"
)

type BigQueryBackend interface {
	CreateDataset(ctx context.Context, project string, ds *model.Dataset) (*model.Dataset, error)
	GetDataset(ctx context.Context, project, dataset string) (*model.Dataset, error)
	ListDatasets(ctx context.Context, project string) ([]*model.Dataset, error)
	DeleteDataset(ctx context.Context, project, dataset string) error

	CreateTable(ctx context.Context, project, dataset string, table *model.Table) (*model.Table, error)
	GetTable(ctx context.Context, project, dataset, table string) (*model.Table, error)
	ListTables(ctx context.Context, project, dataset string) ([]*model.Table, error)
	DeleteTable(ctx context.Context, project, dataset, table string) error

	InsertRows(ctx context.Context, project, dataset, table string, rows []*model.Row) error
	Query(ctx context.Context, query string) (*model.QueryResponse, error)

	Shutdown() error
}
