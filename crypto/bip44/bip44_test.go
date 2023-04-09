package bip44

import (
	"blockchain-libra/crypto/bip32"
	"blockchain-libra/crypto/bip39"
	"encoding/json"
	"testing"
)

func Test_NewKeyFromMnemonic(t *testing.T) {
	seed, err := bip39.NewSeedWithErrorChecking("gorilla easy one advance lesson name math clog awake private aerobic canvas kidney attend food amazing upper interest chicken shadow hip giraffe food curious", "")
	if err != nil {
		t.Log(err)
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		t.Log(err)
	}

	key, err := json.Marshal(masterKey)
	if err != nil {
		t.Log(err)
	}

	t.Log(string(key))

	xkey, _ := NewKeyFromMasterKey(masterKey, 0, 0, 0, 2, 0, 0)
	t.Log(bip32.JsonString(xkey))

}

func Test_NewKeyFromMasterKey(t *testing.T) {

}
