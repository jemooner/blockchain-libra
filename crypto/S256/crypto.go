package S256

import (
	"bytes"
	"crypto/elliptic"
	"eth-wallet/pkg/crypto/ecdsa"
	"fmt"
	"math/big"
)

// SCrypto S256 crypto struct
type SCrypto struct {
}

const (
	PublicKeyCompressedLength = 33
)

func compressPublicKey(x *big.Int, y *big.Int) []byte {
	var key bytes.Buffer

	// Write header; 0x2 for even y value; 0x3 for odd
	key.WriteByte(byte(0x2) + byte(y.Bit(0)))

	// Write X coord; Pad the key so x is aligned with the LSB. Pad size is key length - header size (1) - xBytes size
	xBytes := x.Bytes()
	for i := 0; i < (PublicKeyCompressedLength - 1 - len(xBytes)); i++ {
		key.WriteByte(0x0)
	}
	key.Write(xBytes)

	return key.Bytes()
}

// GetPublicKeyFromPriKey Get the public key from the private key
func (c *SCrypto) GetPublicKeyFromPriKey(priKey []byte) (pubKey []byte, err error) {
	if len(priKey) == 0 {
		return nil, fmt.Errorf("SCrypto GetPublicKeyFromPriKey parameter error")
	}
	curve := ecdsa.S256()
	x, y := curve.ScalarBaseMult(priKey)
	return elliptic.Marshal(curve, x, y), nil
}

// Sign content
func (c *SCrypto) Sign(priKey []byte, plain []byte) (sig []byte, err error) {
	if len(priKey) == 0 || len(plain) == 0 {
		return nil, fmt.Errorf("SCrypto Sign parameter error")
	}
	priObj, err := ecdsa.ToECDSA(priKey, true)
	if err != nil {
		return nil, fmt.Errorf("SCrypto ecdsa.ToECDSA err:%v", err.Error())
	}
	return ecdsa.SignCompact(ecdsa.S256(), priObj, plain, false)
}
