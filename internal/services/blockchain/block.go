package blockchain

import (
	"bytes"
	"encoding/gob"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0}
	pow := NewProof(block)

	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block
}

func (b *Block) CreateNextBlock(data string) *Block {
	block := &Block{[]byte{}, []byte(data), b.Hash, 0}
	pow := NewProof(block)

	nonce, hash := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	ErrHandle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	ErrHandle(err)

	return &block
}

func ErrHandle(err error) {
	if err != nil {
		panic("error occurred: " + err.Error())
	}
}
