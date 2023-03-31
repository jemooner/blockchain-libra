package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"eth-wallet/pkg/crypto/bip39"
	"eth-wallet/pkg/crypto/ecdsa"

	"golang.org/x/crypto/pbkdf2"
)

const (
	BitcoinSeed = "Bitcoin seed"
	Mnemonic    = "search crime conversation tag directory joke leaf express interest password"
)

// NewMnemonic 生成助记词
func NewMnemonic() string {
	randm := rand.Reader
	randBytes := make([]byte, 16)
	randm.Read(randBytes)
	mnemonic, err := bip39.NewMnemonic(randBytes)
	if err != nil {
		return ""
	} else {
		return mnemonic
	}
}

// NewKey 助记词生成公私钥
func NewMasterKey(password string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	seed := pbkdf2.Key([]byte(Mnemonic), []byte("mnemonic"+password), 2048, 64, sha512.New)
	hmac := hmac.New(sha512.New, []byte(BitcoinSeed))
	_, err := hmac.Write([]byte(seed))
	if err != nil {
		return nil, nil
	}
	intermediary := hmac.Sum(nil)
	keyBytes := intermediary[:32] // 私钥
	return ecdsa.PrivKeyFromBytes(ecdsa.S256(), keyBytes)
}

// NewAccount 生成新地址
func NewAccount(password string) string {
	seed := pbkdf2.Key([]byte(Mnemonic), []byte("mnemonic"+password), 2048, 64, sha512.New)
	hmac := hmac.New(sha512.New, []byte(BitcoinSeed))
	_, err := hmac.Write([]byte(seed))
	if err != nil {
		return ""
	}
	intermediary := hmac.Sum(nil)
	keyBytes := intermediary[:32] // 私钥
	_, pub := ecdsa.PrivKeyFromBytes(ecdsa.S256(), keyBytes)
	return pub.ToAddress().String()
}
