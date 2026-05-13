package backend

import (
	"context"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/kms/model"
)

type KMSBackend interface {
	CreateKeyRing(ctx context.Context, project, location, name string) (*model.KeyRing, error)
	GetKeyRing(ctx context.Context, name string) (*model.KeyRing, error)
	ListKeyRings(ctx context.Context, project, location string) ([]*model.KeyRing, error)
	DeleteKeyRing(ctx context.Context, name string) error

	CreateCryptoKey(ctx context.Context, parent string, key *model.CryptoKey) (*model.CryptoKey, error)
	GetCryptoKey(ctx context.Context, name string) (*model.CryptoKey, error)
	ListCryptoKeys(ctx context.Context, parent string) ([]*model.CryptoKey, error)
	UpdateCryptoKey(ctx context.Context, name string, key *model.CryptoKey) (*model.CryptoKey, error)

	CreateCryptoKeyVersion(ctx context.Context, parent string) (*model.CryptoKeyVersion, error)
	GetCryptoKeyVersion(ctx context.Context, name string) (*model.CryptoKeyVersion, error)
	ListCryptoKeyVersions(ctx context.Context, parent string) ([]*model.CryptoKeyVersion, error)
	DestroyCryptoKeyVersion(ctx context.Context, name string) (*model.CryptoKeyVersion, error)
	RestoreCryptoKeyVersion(ctx context.Context, name string) (*model.CryptoKeyVersion, error)

	Encrypt(ctx context.Context, name string, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, name string, ciphertext []byte) ([]byte, error)

	Shutdown() error
}
