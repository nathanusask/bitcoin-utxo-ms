package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

//go:generate mockgen -source=./interface.go -destination=mocks/interface_mock.go -package=mongo
type Interface interface {
	InsertMany(ctx context.Context, utxos []*UTXO) error
	ListCoinsForAddress(ctx context.Context, address string) ([]*UTXO, error)
	GetMaxHeight(ctx context.Context) int
	DeleteMany(ctx context.Context, uniqueKeys []bson.M) error
}
