package blocks

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/frankh/rai"
	"github.com/frankh/rai/address"
	"github.com/frankh/rai/uint128"
	"github.com/frankh/rai/utils"
	"github.com/golang/crypto/blake2b"
	"strings"
	// We've forked golang's ed25519 implementation
	// to use blake2b instead of sha3
	"github.com/frankh/crypto/ed25519"
)

const LiveGenesisBlockHash rai.BlockHash = "991CF190094C00F0B68E2E5F75F6BEE95A2E0BD93CEAA4A6734DB9F19B728948"

var live_genesis_block = JsonBlock([]byte(`{
	"type":           "open",
	"source":         "E89208DD038FBB269987689621D52292AE9C35941A7484756ECCED92A65093BA",
	"representative": "xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3",
	"account":        "xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3",
	"work":           "62f05417dd3fb691",
	"signature":      "9F0C933C8ADE004D808EA1985FA746A7E95BA2A38F867640F53EC8F180BDFE9E2C1268DEAD7C2664F356E37ABA362BC58E46DBA03E523A7B5A19E4B6EB12BB02"
}`))

const publish_threshold = 0xffffffc000000000

type BlockType string

const (
	Open    BlockType = "open"
	Receive           = "receive"
	Send              = "send"
	Change            = "change"
)

type JustType struct {
	Type BlockType `json:"type"`
}

type OpenBlock struct {
	Type           BlockType     `json:"type"`
	Source         rai.BlockHash `json:"source"`
	Representative rai.Account   `json:"representative"`
	Account        rai.Account   `json:"account"`
	Work           rai.Work      `json:"work"`
	Signature      rai.Signature `json:"signature"`
}

type SendBlock struct {
	Type        BlockType       `json:"type"`
	Previous    rai.BlockHash   `json:"previous"`
	Balance     uint128.Uint128 `json:"balance"`
	Destination rai.Account     `json:"destination"`
	Work        rai.Work        `json:"work"`
	Signature   rai.Signature   `json:"signature"`
}

type ReceiveBlock struct {
	Type      BlockType     `json:"type"`
	Previous  rai.BlockHash `json:"previous"`
	Source    rai.BlockHash `json:"source"`
	Work      rai.Work      `json:"work"`
	Signature rai.Signature `json:"signature"`
}

type ChangeBlock struct {
	Type           BlockType     `json:"type"`
	Previous       rai.BlockHash `json:"previous"`
	Representative rai.Account   `json:"representative"`
	Work           rai.Work      `json:"work"`
	Signature      rai.Signature `json:"signature"`
}

type Block struct {
	Type           BlockType
	block          interface{}
	Source         rai.BlockHash
	Representative rai.Account
	Account        rai.Account
	Work           rai.Work
	Signature      rai.Signature
	Previous       rai.BlockHash
	Balance        uint128.Uint128
	Destination    rai.Account
}

func JsonBlock(b []byte) Block {
	var t JustType
	json.Unmarshal(b, &t)
	switch t.Type {
	case Open:
		var o OpenBlock
		json.Unmarshal(b, &o)
		return NewBlock(o)
	default:
		panic("Unknown block type! " + t.Type)
	}
}

func NewBlock(block OpenBlock) Block {
	return Block{
		Open,
		block,
		block.Source,
		block.Representative,
		block.Account,
		block.Work,
		block.Signature,
		"",
		uint128.FromInts(0, 0),
		"",
	}
}

func (b Block) Hash() (result []byte) {
	switch b.Type {
	case Open:
		block := b.block.(OpenBlock)
		return HashOpen(block.Source, block.Representative, block.Account)
	case Send:
		block := b.block.(SendBlock)
		return HashSend(block.Previous, block.Destination, block.Balance)
	case Receive:
		block := b.block.(ReceiveBlock)
		return HashReceive(block.Previous, block.Source)
	case Change:
		block := b.block.(ChangeBlock)
		return HashChange(block.Previous, block.Representative)
	default:
		panic("Unknown block type! " + b.Type)
	}
}

func (b Block) HashToString() (result rai.BlockHash) {
	return rai.BlockHash(strings.ToUpper(hex.EncodeToString(b.Hash())))
}

func SignMessage(private_key string, message []byte) (signature []byte) {
	_, priv := address.KeypairFromPrivateKey(private_key)
	return ed25519.Sign(priv, message)
}

func HashBytes(inputs ...[]byte) (result []byte) {
	hash, err := blake2b.New(32, nil)
	if err != nil {
		panic("Unable to create hash")
	}

	for _, b := range inputs {
		hash.Write(b)
	}

	return hash.Sum(nil)
}

func HashReceive(previous rai.BlockHash, source rai.BlockHash) (result []byte) {
	previous_bytes, _ := hex.DecodeString(string(previous))
	source_bytes, _ := hex.DecodeString(string(source))
	return HashBytes(previous_bytes, source_bytes)
}

func HashChange(previous rai.BlockHash, representative rai.Account) (result []byte) {
	previous_bytes, _ := hex.DecodeString(string(previous))
	repr_bytes, _ := address.AddressToPub(representative)
	return HashBytes(previous_bytes, repr_bytes)
}

func HashSend(previous rai.BlockHash, destination rai.Account, balance uint128.Uint128) (result []byte) {
	previous_bytes, _ := hex.DecodeString(string(previous))
	dest_bytes, _ := address.AddressToPub(destination)
	balance_bytes := balance.GetBytes()

	return HashBytes(previous_bytes, dest_bytes, balance_bytes)
}

func HashOpen(source rai.BlockHash, representative rai.Account, account rai.Account) (result []byte) {
	source_bytes, _ := hex.DecodeString(string(source))
	repr_bytes, _ := address.AddressToPub(representative)
	account_bytes, _ := address.AddressToPub(account)
	return HashBytes(source_bytes, repr_bytes, account_bytes)
}

func ValidateWork(block_hash []byte, work []byte) bool {
	hash, err := blake2b.New(8, nil)
	if err != nil {
		panic("Unable to create hash")
	}

	hash.Write(utils.Reversed(work))
	hash.Write(block_hash)

	work_value := hash.Sum(nil)
	work_value_int := binary.LittleEndian.Uint64(work_value)

	return work_value_int >= publish_threshold
}
