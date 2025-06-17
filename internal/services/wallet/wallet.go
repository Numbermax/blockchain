package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w *Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	fmt.Printf("pub key: %x\n", w.PublicKey)
	fmt.Printf("pub hash: %x\n", pubHash)
	fmt.Printf("address: %x\n", address)

	return address
}

func ValidateAddress(address string) bool {
	publicKeyHash := Base58Decode([]byte(address))
	actualChecksum := publicKeyHash[len(publicKeyHash)-checksumLength:]
	version := publicKeyHash[0]
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, publicKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic()
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()

	return &Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	secondHash := sha256.Sum256(pubHash[:])

	// Take first 20 bytes to match RIPEMD160 output size
	return secondHash[:20]
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	checksum := secondHash[:checksumLength]

	return checksum
}

func bytesToPrivateKey(privateKeyBytes []byte) *ecdsa.PrivateKey {
	curve := elliptic.P256()

	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
		},
		D: new(big.Int).SetBytes(privateKeyBytes),
	}

	// not need to regenerate, public key is saved to file
	// if privateKey.PublicKey.X == nil || privateKey.PublicKey.Y == nil {
	// 	fmt.Println("Public key is nil, regenerating...")
	// 	privateKey.PublicKey.X, privateKey.PublicKey.Y = curve.ScalarBaseMult(privateKeyBytes)
	// }

	return privateKey
}
