package common

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// IntToHex int转16进制
func IntToHex(i interface{}) string {
	hex := fmt.Sprintf("%x", i)
	if strings.HasPrefix(hex, "0x") {
		return hex
	}
	return "0x" + hex
}

func HexToUnit64(hex string) (number uint64, err error) {
	if len(hex) < 2 {
		return 0, nil
	}
	return strconv.ParseUint(hex[2:], 16, 64)
}

func HexToInt64(hex string) (number int64, err error) {
	return strconv.ParseInt(hex[2:], 16, 64)
}

// HexToBigInt
func HexToBigInt(hex string) *big.Int {
	if strings.HasPrefix(hex, "0x") {
		hex = hex[2:]
	}
	n := new(big.Int)
	n, _ = n.SetString(hex[:], 16)
	return n
}

// BigIntToHex
func BigIntToHex(b *big.Int) string {
	if b == nil {
		return ""
	}
	return hex.EncodeToString(b.Bytes())
}

// GetRandString 随机生成N位字符串
func GetRandString(n int) string {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(mainBuff)[:n]
}

func ClearZero(v string) (res string) {
	if v == "" {
		return ""
	}
	rule, _ := regexp.Compile("^0+")
	res = rule.ReplaceAllString(v[2:], "")
	return
}
func ClearBackZero(v string) (res string) {
	if v == "" {
		return ""
	}
	rule, _ := regexp.Compile("0+$")
	res = rule.ReplaceAllString(v, "")
	return
}
func FillZero(s string) string {
	for i := len(s); i < 64; i++ {
		s = "0" + s
	}
	return s
}
func FillZero256(s string) string {
	for i := len(s); i < 256; i++ {
		s = "0" + s
	}
	return s
}
func HexToAmount(hex string) (number string, err error) {
	hex = hex[2:]
	if len(hex) != 64 {
		return "", errors.New("length not equal 64")
	}
	valueByte := hex[:64]
	if err != nil {
		return "", errors.New("HexToInt64 value err")
	}
	return valueByte, nil
}

func HexToString(hex string) (number string, err error) {
	hex = hex[2:]
	if len(hex) != 192 {
		return "", errors.New("length not equal 192")
	}
	allSizeByte := hex[:64]
	_, err = HexToInt64(allSizeByte)
	if err != nil {
		return "", errors.New("HexToInt64 allSize err")
	}
	sizeByte := hex[64:128]
	size, err := HexToInt64(sizeByte)
	if err != nil {
		return "", errors.New("HexToInt64 size err")
	}
	valueByte := hex[128:192]
	if err != nil {
		return "", errors.New("HexToInt64 value err")
	}
	return valueByte[:size*2], nil
}
