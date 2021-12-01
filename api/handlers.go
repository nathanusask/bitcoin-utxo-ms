package api

import (
	"net/http"

	_mongo "github.com/ABMatrix/bitcoin-utxo-ms/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	KEY_ERROR = "error"
)

type ListPayload struct {
	Address string `json:"address"`
}

type Server struct {
	utxoCollection *mongo.Collection
}

func New(mongoCli *mongo.Client, db string, collection string) *Server {
	return &Server{utxoCollection: mongoCli.Database(db).Collection(collection)}
}

func (s Server) ListHandler(c *gin.Context) {
	payload := &ListPayload{}
	if err := c.BindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{KEY_ERROR: err.Error()})
		return
	}
	if payload.Address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]string{KEY_ERROR: "address cannot by empty"})
		return
	}

	filter := bson.M{_mongo.KEY_ADDRESS: payload.Address, _mongo.KEY_AMOUNT: bson.M{_mongo.KEY_GT: 0}}
	cur, err := s.utxoCollection.Find(c, filter)
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

	c.JSON(http.StatusOK, utxos)
}
