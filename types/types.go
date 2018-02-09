package types

import (
	"encoding/hex"
	"strings"

	"github.com/frankh/crypto/ed25519"
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

func (sig Signature) ToBytes() []byte {
	bytes, err := hex.DecodeString(string(sig))
	if err != nil {
		panic(err)
	}
	return bytes
}

func (hash BlockHash) Sign(private_key ed25519.PrivateKey) Signature {
	sig := hex.EncodeToString(ed25519.Sign(private_key, hash.ToBytes()))
	return Signature(strings.ToUpper(sig))
}

func BlockHashFromBytes(b []byte) BlockHash {
	return BlockHash(strings.ToUpper(hex.EncodeToString(b)))
}
