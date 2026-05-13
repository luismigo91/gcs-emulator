package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/iam/model"
)

type IAMBackend interface {
	CreateServiceAccount(ctx context.Context, project string, sa *model.ServiceAccount) (*model.ServiceAccount, error)
	GetServiceAccount(ctx context.Context, name string) (*model.ServiceAccount, error)
	ListServiceAccounts(ctx context.Context, project string) ([]*model.ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, name string) error
	SignJwt(ctx context.Context, name string, payload string) (string, error)
	GenerateAccessToken(ctx context.Context, name string) (*model.GenerateAccessTokenResponse, error)
	Shutdown() error
}
