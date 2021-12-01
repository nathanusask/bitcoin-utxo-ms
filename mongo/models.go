package mongo

type ScriptType string

const (
	ScriptType_P2PK        ScriptType = "p2pk"
	ScriptType_P2PKH       ScriptType = "p2pkh"
	ScriptType_P2SH        ScriptType = "p2sh"
	ScriptType_P2WKH       ScriptType = "p2wkh"
	ScriptType_P2WSH       ScriptType = "p2wsh"
	ScriptType_NonStandard ScriptType = "non-standard"
)

type UTXO struct {
	TxID     string     `json:"tx_id" bson:"tx_id"`
	Vout     int        `json:"vout" bson:"vout"`
	Height   int        `json:"height" bson:"height"`
	Coinbase bool       `json:"coinbase" bson:"coinbase"`
	Amount   int64      `json:"amount" bson:"amount"`
	Size     int64      `json:"size" bson:"size"`
	Script   string     `json:"script" bson:"script"`
	Type     ScriptType `json:"type" bson:"type"` // TODO: maybe not string?
	Address  string     `json:"address" bson:"address"`
}
