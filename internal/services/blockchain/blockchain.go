package blockchain

import (
	"log/slog"
)

type BlockchainService struct {
	logger *slog.Logger
}

type BlockChain struct {
	blocks []*Block
}

func NewBlockchainService(logger *slog.Logger) *BlockchainService {
	return &BlockchainService{
		logger: logger,
	}
}

func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := prevBlock.CreateNextBlock(data)
	chain.blocks = append(chain.blocks, new)
}

func (chain *BlockChain) GetBlocks() []*Block {
	return chain.blocks
}

func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
