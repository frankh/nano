package rai

import (
	"encoding/hex"
	"strings"
)

type BlockHash string
type Account string
type Work string
type Signature string

func (hash BlockHash) ToBytes() []byte {
	bytes, err := hex.DecodeString(string(hash))
	if err != nil {
		panic(err)
	}
	return bytes
}

func BlockHashFromBytes(b []byte) BlockHash {
	return BlockHash(strings.ToUpper(hex.EncodeToString(b)))
}
