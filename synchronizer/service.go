package synchronizer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/ABMatrix/bitcoin-utxo-ms/fullnode"
	"github.com/ABMatrix/bitcoin-utxo-ms/mongo"
)

const (
	INTERVAL = 2 * time.Minute
)

type server struct {
	mongoServer mongo.Interface
	fullnode    fullnode.Interface
	wg          *sync.WaitGroup
}

var MapMongoScriptType2BlockScriptType = map[mongo.ScriptType]fullnode.ScriptType{
	mongo.ScriptType_P2PK:        fullnode.ScriptType_P2PK,
	mongo.ScriptType_P2PKH:       fullnode.ScriptType_P2PKH,
	mongo.ScriptType_P2SH:        fullnode.ScriptType_P2SH,
	mongo.ScriptType_P2WKH:       fullnode.ScriptType_P2WKH,
	mongo.ScriptType_P2WSH:       fullnode.ScriptType_P2WSH,
	mongo.ScriptType_NonStandard: fullnode.ScriptType_NonStandard,
}

var MapBlockScriptType2MongoScriptType = map[fullnode.ScriptType]mongo.ScriptType{
	fullnode.ScriptType_P2PK:        mongo.ScriptType_P2PK,
	fullnode.ScriptType_P2PKH:       mongo.ScriptType_P2PKH,
	fullnode.ScriptType_P2SH:        mongo.ScriptType_P2SH,
	fullnode.ScriptType_P2WKH:       mongo.ScriptType_P2WKH,
	fullnode.ScriptType_P2WSH:       mongo.ScriptType_P2WSH,
	fullnode.ScriptType_NonStandard: mongo.ScriptType_NonStandard,
}

func New(m mongo.Interface, f fullnode.Interface) Interface {
	return &server{
		mongoServer: m,
		fullnode:    f,
		wg:          &sync.WaitGroup{},
	}
}

func (s server) Start(ctx context.Context) {
	maxHeightInDatabase := s.mongoServer.GetMaxHeight(ctx)
	if maxHeightInDatabase < 0 {
		log.Println("failed to get max height from database")
		return
	}

	height := maxHeightInDatabase + 1
	s.syncBlockStartingAtHeight(ctx, &height)
	log.Println("[debug] all blocks have been synced before ", height)

	ticker := time.NewTicker(INTERVAL)
	go func(initialHeight int) {
		currentHeight := initialHeight
		for {
			select {
			case <-ticker.C:
				height := s.fullnode.GetBestBlockHeight()
				if height >= currentHeight {
					s.syncBlockStartingAtHeight(ctx, &currentHeight)
				}
			}
		}
	}(height)
}

// syncBlockStartingAtHeight is a blocking method
// the returned `height` is the height of the next block which hasn't arrived yet
func (s server) syncBlockStartingAtHeight(ctx context.Context, height *int) {
	curBlock := s.fullnode.GetBlockAtHeight(*height)
	for curBlock != nil {
		// sync one block at a time
		s.wg.Add(1)
		go s.syncOneBlock(ctx, curBlock, true)
		*height++

		// getting the next block takes some time and can run simultaneously alongside block syncing
		if curBlock.NextBlockHash != "" {
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				curBlock = s.fullnode.GetBlock(curBlock.NextBlockHash)
			}()
			s.wg.Wait()
			continue
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			curBlock = s.fullnode.GetBlockAtHeight(*height)
		}()
		s.wg.Wait()
	}
}

func (s server) syncOneBlock(ctx context.Context, block *fullnode.Block, needWg bool) {
	if needWg {
		defer s.wg.Done()
	}
	if block == nil {
		return
	}
	log.Println(fmt.Sprintf("[debug] syncing block at height %d...", block.Height))
	defer log.Println(fmt.Sprintf("[debug] exiting syncing block at height %d...", block.Height))
	var deleteKeys []bson.M
	var insertUtxos []*mongo.UTXO
	for index, transaction := range block.Transactions {
		if transaction == nil {
			continue
		}
		if index == 0 {
			// coinbase transaction
			for _, txout := range transaction.TxOuts {
				if txout == nil {
					// we don't give a crap about if the coin is spendable
					continue
				}
				insertUtxos = append(insertUtxos, &mongo.UTXO{
					TxID:     transaction.Txid,
					Vout:     txout.Index,
					Height:   block.Height,
					Coinbase: true, // the only difference lies here
					Amount:   int64(txout.Value * 1e8),
					Size:     0,
					Script:   txout.Script.Hex,
					Type:     MapBlockScriptType2MongoScriptType[txout.Script.Type],
					Address:  txout.Script.Address,
				})
			}
		}

		for _, txin := range transaction.TxIns {
			if txin == nil {
				continue
			}
			deleteKeys = append(deleteKeys, bson.M{mongo.KEY_TXID: txin.Txid, mongo.KEY_VOUT: txin.Vout})
		}
		for _, txout := range transaction.TxOuts {
			if txout == nil {
				// we don't give a crap about if the coin is spendable
				continue
			}
			insertUtxos = append(insertUtxos, &mongo.UTXO{
				TxID:     transaction.Txid,
				Vout:     txout.Index,
				Height:   block.Height,
				Coinbase: false,
				Amount:   int64(txout.Value * 1e8),
				Size:     0,
				Script:   txout.Script.Hex,
				Type:     MapBlockScriptType2MongoScriptType[txout.Script.Type],
				Address:  txout.Script.Address,
			})
		}
	}
	wgInside := &sync.WaitGroup{}
	wgInside.Add(1)
	go func() {
		defer wgInside.Done()

		if err := s.mongoServer.DeleteMany(ctx, deleteKeys); err != nil {
			log.Println("[error] failed to delete many with error: ", err.Error())
		}
	}()

	wgInside.Add(1)
	go func() {
		defer wgInside.Done()

		if err := s.mongoServer.InsertMany(ctx, insertUtxos); err != nil {
			log.Println("[error] failed to insert many with error: ", err.Error())
		}
	}()

	wgInside.Wait()

	log.Println(fmt.Sprintf("[debug] successfully finished syncing block at height %d...", block.Height))
}
