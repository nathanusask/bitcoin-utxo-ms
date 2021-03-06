package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ABMatrix/bitcoin-utxo-ms/api"
	"github.com/ABMatrix/bitcoin-utxo-ms/middleware"

	"github.com/ABMatrix/bitcoin-utxo-ms/synchronizer"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ABMatrix/bitcoin-utxo-ms/fullnode"
	_mongo "github.com/ABMatrix/bitcoin-utxo-ms/mongo"

	"github.com/gin-gonic/gin"
)

const (
	ENV_BTC_FULL_NODE_URI    = "BTC_FULL_NODE_URI"
	ENV_BTC_DATABASE_NAME    = "BTC_DATABASE_NAME"
	ENV_UTXO_COLLECTION_NAME = "UTXO_COLLECTION_NAME"
	ENV_MONGO_URI            = "MONGO_URI"
	ENV_PORT                 = "PORT"
)

func main() {
	btcUri := os.Getenv(ENV_BTC_FULL_NODE_URI)
	if btcUri == "" {
		log.Fatalln(ENV_BTC_FULL_NODE_URI, " is unset!")
	}

	mongoUri := os.Getenv(ENV_MONGO_URI)
	if mongoUri == "" {
		log.Fatalln(ENV_MONGO_URI, " is unset!")
	}

	btcDatabase := os.Getenv(ENV_BTC_DATABASE_NAME)
	if btcDatabase == "" {
		log.Fatalln(ENV_BTC_DATABASE_NAME, " is unset!")
	}

	utxoCollection := os.Getenv(ENV_UTXO_COLLECTION_NAME)
	if utxoCollection == "" {
		log.Fatalln(ENV_UTXO_COLLECTION_NAME, " is unset!")
	}

	port := os.Getenv(ENV_PORT)
	if port == "" {
		log.Fatalln(ENV_PORT, " is unset!")
	}

	ctx := context.Background()

	// initialize mongodb
	clientOptions := options.Client().ApplyURI(mongoUri)

	// connect to MongoDB
	mongoCli, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalln("[error] failed to connect mongo with error:", err)
	}
	// check connection
	err = mongoCli.Ping(ctx, nil)
	if err != nil {
		log.Fatalln("[error] failed to ping mongodb with error: ", err)
	}

	log.Println("[debug] mongo connection is OK...")

	btcServer := fullnode.New(btcUri)
	mongoServer := _mongo.New(mongoCli, btcDatabase, utxoCollection)
	syncer := synchronizer.New(mongoServer, btcServer)
	syncer.Start(ctx) // this is a blocking process!!

	// initialize gin web server
	router := gin.Default()
	router.Use(middleware.Cors())

	apiServer := api.New(mongoCli, btcDatabase, utxoCollection)
	utxoQuery := router.Group("/utxo")
	utxoQuery.POST("list", apiServer.ListHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}
	log.Fatalln(httpServer.ListenAndServe()) // TODO: change it to HTTPS server
}
