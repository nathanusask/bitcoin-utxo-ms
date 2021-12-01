package fullnode

//go:generate mockgen -source=./interface.go -destination=mocks/interface_mock.go -package=fullnode
type Interface interface {
	GetBestBlockHeight() int
	GetBlockAtHeight(height int) *Block
	GetBestBlock() *Block
	GetBlock(hash string) *Block
}
