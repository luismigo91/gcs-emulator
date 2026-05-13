package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/dns/model"
)

type DNSBackend interface {
	CreateZone(ctx context.Context, project string, zone *model.ManagedZone) (*model.ManagedZone, error)
	GetZone(ctx context.Context, project, name string) (*model.ManagedZone, error)
	ListZones(ctx context.Context, project string) ([]*model.ManagedZone, error)
	DeleteZone(ctx context.Context, project, name string) error
	CreateChange(ctx context.Context, project, zone string, change *model.Change) ([]*model.ResourceRecordSet, error)
	ListRecordSets(ctx context.Context, project, zone string) ([]*model.ResourceRecordSet, error)
	Shutdown() error
}
