package mongo

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	KEY_ADDRESS = "address"
	KEY_HEIGHT  = "height"
	KEY_TXID    = "tx_id"
	KEY_VOUT    = "vout"
	KEY_AMOUNT  = "amount"
	KEY_GT      = "$gt"

	ENV_MONGO_UTXO_KEY_INDEX_NAME = "MONGO_UTXO_KEY_INDEX_NAME"
)

type server struct {
	collection *mongo.Collection
}

func New(c *mongo.Client, db string, collection string) Interface {
	return &server{
		collection: c.Database(db).Collection(collection),
	}
}

func (s server) InsertMany(ctx context.Context, utxos []*UTXO) error {
	var documents []interface{}
	for _, utxo := range utxos {
		documents = append(documents, utxo)
	}
	_, err := s.collection.InsertMany(ctx, documents)
	return err
}

func (s server) ListCoinsForAddress(ctx context.Context, address string) ([]*UTXO, error) {
	filter := bson.M{
		KEY_ADDRESS: address,
	}

	cur, err := s.collection.Find(ctx, filter)
	if err != nil {
		log.Println(fmt.Sprintf("failed to find documents for %s with error: %s", address, err.Error()))
		return nil, err
	}

	var utxos []*UTXO
	for cur.Next(ctx) {
		utxo := &UTXO{}
		if err := bson.Unmarshal(cur.Current, &utxo); err != nil {
			log.Println("failed to unmarshal utxo with error: ", err.Error())
			return nil, err
		}
		utxos = append(utxos, utxo)
	}

	return utxos, nil
}

func (s server) GetMaxHeight(ctx context.Context) int {
	cur := s.collection.FindOne(ctx, bson.M{}, &options.FindOneOptions{Sort: bson.M{KEY_HEIGHT: -1}})
	utxo := &UTXO{}
	if err := cur.Decode(&utxo); err != nil {
		log.Println("failed to decode utxo with error: ", err.Error())
		return -1
	}
	return utxo.Height
}

func (s server) DeleteMany(ctx context.Context, uniqueKeys []bson.M) error {
	var writeModels []mongo.WriteModel
	for _, key := range uniqueKeys {
		deleteModel := &mongo.DeleteManyModel{
			Filter: key,
		}
		if utxoKeyIndex := os.Getenv(ENV_MONGO_UTXO_KEY_INDEX_NAME); utxoKeyIndex != "" {
			deleteModel.SetHint(utxoKeyIndex)
		}
		writeModels = append(writeModels, deleteModel)
	}

	_, err := s.collection.BulkWrite(ctx, writeModels)
	return err
}
