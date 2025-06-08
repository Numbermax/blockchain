package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

const walletFile = "./tmp/wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

// SerializableWallet for gob encoding/decoding
type SerializableWallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadWallet()

	return &wallets, err
}

func (w *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range w.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (w *Wallets) GetWallet(address string) Wallet {
	return *w.Wallets[address]
}

func (ws *Wallets) SaveFile() {
	var content bytes.Buffer

	// Convert wallets to serializable format
	serializableWallets := make(map[string]*SerializableWallet)
	for address, wallet := range ws.Wallets {
		serializableWallets[address] = &SerializableWallet{
			PrivateKey: wallet.PrivateKey.D.Bytes(), // Convert big.Int to bytes
			PublicKey:  wallet.PublicKey,
		}
	}

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(serializableWallets)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic()
	}
}

func (w *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())

	w.Wallets[address] = wallet

	return address
}

func (ws *Wallets) LoadWallet() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		fmt.Println("No wallet file found")
		return err
	}

	var serializableWallets map[string]*SerializableWallet

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&serializableWallets)
	if err != nil {
		return err
	}

	// Convert back to wallets
	ws.Wallets = make(map[string]*Wallet)
	for address, serWallet := range serializableWallets {
		// Reconstruct the private key from bytes
		privateKey := bytesToPrivateKey(serWallet.PrivateKey)
		ws.Wallets[address] = &Wallet{
			PrivateKey: *privateKey,
			PublicKey:  serWallet.PublicKey,
		}
	}

	return nil
}
