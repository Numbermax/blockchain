package blockchain

import (
	"fmt"
	"log/slog"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blockchain.db"
	lastHashKey = "lh"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lastHashKey))
		ErrHandle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		ErrHandle(err)
		return nil
	})
	new := CreateBlock(data, lastHash)
	ErrHandle(err)
	slog.Info("Adding new block", slog.String("hash", fmt.Sprintf("%x", new.Hash)), slog.String("data", string(new.Data)))
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(new.Hash, new.Serialize())
		ErrHandle(err)
		chain.LastHash = new.Hash
		return nil
	})
	ErrHandle(err)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(lastHashKey), new.Hash)
		ErrHandle(err)
		return err
	})
	ErrHandle(err)
}

func InitBlockChain() *BlockChain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	ErrHandle(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte(lastHashKey)); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found, creating genesis block...")
			genesis := Genesis()
			fmt.Println("Genesis block created with hash:", genesis.Hash)
			err = txn.Set(genesis.Hash, genesis.Serialize())
			ErrHandle(err)
			lastHash = genesis.Hash
			err = txn.Set([]byte(lastHashKey), genesis.Hash)

			return err
		} else {
			item, err := txn.Get([]byte(lastHashKey))
			if err != nil {
				return err
			}
			err = item.Value(func(val []byte) error {
				lastHash = val
				return nil
			})
			return err
		}
	})

	ErrHandle(err)

	return &BlockChain{
		LastHash: lastHash,
		Database: db,
	}
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		ErrHandle(err)

		err = item.Value(func(val []byte) error {
			block = Deserialize(val)

			return nil
		})

		return err
	})

	ErrHandle(err)

	iter.CurrentHash = block.PrevHash

	return block
}
