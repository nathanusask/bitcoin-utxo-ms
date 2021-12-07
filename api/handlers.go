package api

import (
	"math"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo/options"

	_mongo "github.com/ABMatrix/bitcoin-utxo-ms/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Order int

const (
	OrderAsc  Order = 1
	OrderDesc Order = -1
)

const (
	KEY_ERROR          = "error"
	KEY_AMOUNT         = "amount"
	DefaultPage  int64 = 1
	DefaultLimit int64 = 20
)

type ListRequest struct {
	Address string `json:"address,omitempty"`
	Page    int64  `json:"page,omitempty"`
	Limit   int64  `json:"limit,omitempty"`
	Order   Order  `json:"order,omitempty"`
}

type ListResponse struct {
	UTXOS    []*_mongo.UTXO `json:"utxos,omitempty"`
	Total    int64          `json:"total,omitempty"`
	Page     int64          `json:"page,omitempty"`
	LastPage int64          `json:"last_page,omitempty"`
}

type Server struct {
	utxoCollection *mongo.Collection
}

func New(mongoCli *mongo.Client, db string, collection string) *Server {
	return &Server{utxoCollection: mongoCli.Database(db).Collection(collection)}
}

func (s Server) ListHandler(c *gin.Context) {
	payload := &ListRequest{}
	if err := c.BindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{KEY_ERROR: err.Error()})
		return
	}
	if payload.Address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{KEY_ERROR: "address cannot by empty"})
		return
	}

	filter := bson.M{_mongo.KEY_ADDRESS: payload.Address, _mongo.KEY_AMOUNT: bson.M{_mongo.KEY_GT: 0}}
	var findOption *options.FindOptions
	var page = DefaultPage
	var limit = DefaultLimit
	if payload.Page > 0 {
		page = payload.Page
	}
	if payload.Limit > 0 {
		limit = payload.Limit
	}
	findOption.SetSkip((page - 1) * limit)

	var sortOrder Order = OrderAsc
	if payload.Order == OrderAsc || payload.Order == OrderDesc {
		sortOrder = payload.Order
	}
	findOption.SetSort(bson.M{KEY_AMOUNT: sortOrder})
	cur, err := s.utxoCollection.Find(c, filter, findOption)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{KEY_ERROR: err.Error()})
		return
	}

	total, err := s.utxoCollection.CountDocuments(c, filter)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{KEY_ERROR: err.Error()})
		return
	}

	var utxos []*_mongo.UTXO
	for cur.Next(c) {
		utxo := &_mongo.UTXO{}
		if err := cur.Decode(&utxo); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{KEY_ERROR: err.Error()})
			return
		}
		utxos = append(utxos, utxo)
	}

	c.JSON(http.StatusOK, ListResponse{
		UTXOS:    utxos,
		Total:    total,
		Page:     page,
		LastPage: int64(math.Ceil(float64(total / limit))),
	})
}
