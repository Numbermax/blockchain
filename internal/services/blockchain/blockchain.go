package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	lastHashKey = "lh"
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	logger   slog.Logger
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
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
	new := CreateBlock(transactions, lastHash)
	ErrHandle(err)
	slog.Info("Adding new block", slog.String("hash", fmt.Sprintf("%x", new.Hash)))
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

func (chain *BlockChain) FindUnspentTransactions(pubHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTxOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTxOs[txID] != nil {
					for _, spentOut := range spentTxOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubHash) {
						inTxId := hex.EncodeToString(in.ID)
						spentTxOs[inTxId] = append(spentTxOs[inTxId], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

func (chain *BlockChain) FindUTXO(pubHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(pubHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func InitBlockChain(logger slog.Logger, address string) *BlockChain {
	op := "services.blockchain.blockchain.InitBlockChain"
	var lastHash []byte
	logger.With(slog.String("operation", op))

	if DbExists() {
		logger.Info("Database already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	ErrHandle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		fmt.Println("No existing blockchain found, creating genesis block...")
		genesis := Genesis(cbtx)
		fmt.Println("Genesis block created with hash:", genesis.Hash)
		err = txn.Set(genesis.Hash, genesis.Serialize())
		ErrHandle(err)
		lastHash = genesis.Hash
		err = txn.Set([]byte(lastHashKey), genesis.Hash)

		return err
	})

	ErrHandle(err)

	return &BlockChain{
		LastHash: lastHash,
		Database: db,
	}
}

func ContinueBlockChain(logger slog.Logger, address string) *BlockChain {
	op := "services.blockchain.blockchain.InitBlockChain"
	var lastHash []byte
	logger.With(slog.String("operation", op))

	if !DbExists() {
		fmt.Println("No existing blockchain, create one")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	ErrHandle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lastHashKey))
		ErrHandle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})

		return err
	})

	return &BlockChain{
		logger:   logger,
		LastHash: lastHash,
		Database: db,
	}
}

func DbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
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

func (bc *BlockChain) FindTransaction(Id []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, Id) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		ErrHandle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction, privKey ecdsa.PrivateKey) bool {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		ErrHandle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
