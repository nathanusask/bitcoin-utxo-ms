package fullnode

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type RpcMethods string

const (
	KEY_CONTENT_TYPE        = "Content-Type"
	CONTENT_TYPE_TEXT_PLAIN = "text/plain"

	RpcMethodsGetBestBlockHash RpcMethods = "getbestblockhash"
	RpcMethodsGetBlockHash     RpcMethods = "getblockhash"
	RpcMethodsGetBlock         RpcMethods = "getblock"
	RpcMethodsGetBlockCount    RpcMethods = "getblockcount"
)

type server struct {
	rpcUrl     string
	httpClient *http.Client
}

func New(rpcUrl string) Interface {
	return &server{
		rpcUrl:     rpcUrl,
		httpClient: &http.Client{},
	}
}

type Payload struct {
	JsonRpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  RpcMethods    `json:"method"`
	Params  []interface{} `json:"params"`
}

func (s server) GetBlockAtHeight(height int) *Block {
	payload := &Payload{
		Method: RpcMethodsGetBlockHash,
		Params: []interface{}{height},
	}

	result := s.rpcCall(payload)
	if result == nil {
		log.Println("rpc call to get block hash has failed!")
		return nil
	}

	blockHash, ok := result.(string)
	if !ok {
		log.Println("failed to cast result to string")
		return nil
	}
	log.Printf("block hash at height %d is %s\n", height, blockHash)

	return s.GetBlock(blockHash)
}

func (s server) GetBestBlock() *Block {
	payload := &Payload{
		Method: RpcMethodsGetBestBlockHash,
		Params: []interface{}{},
	}

	result := s.rpcCall(payload)
	if result == nil {
		log.Println("rpc call to get best block hash has failed!")
		return nil
	}

	blockHash, ok := result.(string)
	if !ok {
		log.Println("failed to case result in to string")
		return nil
	}

	return s.GetBlock(blockHash)
}

func (s server) GetBlock(hash string) *Block {
	payload := &Payload{
		Method: RpcMethodsGetBlock,
		Params: []interface{}{hash, 2},
	}

	result := s.rpcCall(payload)
	if result == nil {
		log.Println("rpc call to get best block hash has failed!")
		return nil
	}

	block := &Block{}
	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(result)
	json.NewDecoder(buf).Decode(block)

	return block
}

func (s server) GetBestBlockHeight() int {
	payload := &Payload{
		Method: RpcMethodsGetBlockCount,
		Params: []interface{}{},
	}

	result := s.rpcCall(payload)
	if result == nil {
		log.Println("rpc call to get best block hash has failed!")
		return -1
	}

	heightF, ok := result.(float64)
	if !ok {
		log.Println("failed to cast result to float64 as result is ", result)
		return -1
	}
	return int(heightF)
}

func (s server) rpcCall(payload *Payload) interface{} {
	url := s.rpcUrl

	if payload.JsonRpc == "" {
		payload.JsonRpc = "1.0"
	}

	if payload.ID == "" {
		payload.ID = "1"
	}

	marshaled, err := json.Marshal(payload)
	if err != nil {
		log.Println("failed to marshal payload with error: ", err.Error())
		return nil
	}

	reader := strings.NewReader(string(marshaled))

	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		log.Println("failed to initialize a new request with error: ", err.Error())
		return nil
	}
	req.Header.Add(KEY_CONTENT_TYPE, CONTENT_TYPE_TEXT_PLAIN)

	res, err := s.httpClient.Do(req)
	if err != nil {
		log.Println("failed to send post request with error: ", err.Error())
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("failed to read body with error: ", err.Error())
		return nil
	}

	response := &Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Println("failed to unmarshal response with error: ", err.Error())
		return nil
	}

	if response.Error != nil {
		log.Printf("unsuccessful request:\nerror code: %d\nmessage: %s\n", response.Error.Code, response.Error.Message)
		return nil
	}

	return response.Result
}
