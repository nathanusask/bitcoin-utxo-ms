package fullnode

import "time"

type ScriptType string

const (
	ScriptType_P2PK        ScriptType = "pubkey"
	ScriptType_P2PKH       ScriptType = "pubkeyhash"
	ScriptType_P2SH        ScriptType = "scripthash"
	ScriptType_P2WKH       ScriptType = "witness_v0_keyhash"
	ScriptType_P2WSH       ScriptType = "witness_v0_scripthash"
	ScriptType_NonStandard ScriptType = "nulldata"
)

type TxIn struct {
	Txid        string   `json:"txid"`
	Vout        int64    `json:"vout"`
	ScriptSig   *Script  `json:"scriptSig"`
	Coinbase    string   `json:"coinbase"`
	TxinWitness []string `json:"txinwitness"`
	Sequence    int64    `json:"sequence"`
}

type Script struct {
	Asm     string     `json:"asm"`
	Hex     string     `json:"hex"`
	Address string     `json:"address"`
	Type    ScriptType `json:"type"`
}

type TxOut struct {
	Value  float64 `json:"value"`
	Index  int     `json:"n"`
	Script *Script `json:"scriptPubKey,omitempty"`
}

type Transaction struct {
	Txid     string    `json:"txid"`
	Hash     string    `json:"hash"`
	Version  int64     `json:"version"`
	Size     int64     `json:"size"`
	Vsize    int64     `json:"vsize"`
	Weight   int64     `json:"weight"`
	LockTime time.Time `json:"locktime"`
	TxIns    []*TxIn   `json:"vin"`
	TxOuts   []*TxOut  `json:"vout"`
	Hex      string    `json:"hex"`
	Fee      float64   `json:"fee,omitempty"`
}

type Block struct {
	Hash              string         `json:"hash"`
	Confirmations     int            `json:"confirmations"`
	Height            int            `json:"height"`
	Version           int            `json:"version"`
	VersionHex        string         `json:"versionHex"`
	MerkleRoot        string         `json:"merkleroot"`
	Time              int            `json:"time"`
	MedianTime        int            `json:"mediantime"`
	Nonce             int64          `json:"nonce"`
	Bits              string         `json:"bits"`
	Difficulty        float64        `json:"difficulty"`
	ChainWork         string         `json:"chainwork"`
	NTx               int            `json:"nTx"`
	PreviousBlockHash string         `json:"previousblockhash"`
	NextBlockHash     string         `json:"nextblockhash"`
	StrippedSize      int            `json:"strippedsize"`
	Size              int            `json:"size"`
	Weight            int            `json:"weight"`
	Transactions      []*Transaction `json:"tx"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Result interface{} `json:"result"`
	Error  *Error      `json:"error"`
	ID     string      `json:"id"`
}
