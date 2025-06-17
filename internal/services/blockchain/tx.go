package blockchain

import (
	"bytes"

	"github.com/numbermax/blockchain/internal/services/wallet"
)

type TxOutput struct {
	Value         int
	PublicKeyHash []byte
}

type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PublicKey []byte
}

func (in *TxInput) UsesKey(publicKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PublicKey)

	return bytes.Compare(lockingHash, publicKeyHash) == 0
}

func NewTxOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PublicKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(publicKeyHash []byte) bool {
	return bytes.Compare(out.PublicKeyHash, publicKeyHash) == 0
}
