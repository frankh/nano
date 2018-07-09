package blocks

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"hash"
	"strings"

	"github.com/frankh/nano/address"
	"github.com/frankh/nano/types"
	"github.com/frankh/nano/uint128"
	"github.com/frankh/nano/utils"
	"github.com/golang/crypto/blake2b"
	// We've forked golang's ed25519 implementation
	// to use blake2b instead of sha3
	"github.com/frankh/crypto/ed25519"
)

const LiveGenesisBlockHash types.BlockHash = "991CF190094C00F0B68E2E5F75F6BEE95A2E0BD93CEAA4A6734DB9F19B728948"
const LiveGenesisSourceHash types.BlockHash = "E89208DD038FBB269987689621D52292AE9C35941A7484756ECCED92A65093BA"

var GenesisAmount uint128.Uint128 = uint128.FromInts(0xffffffffffffffff, 0xffffffffffffffff)
var WorkThreshold = uint64(0xffffffc000000000)

const TestPrivateKey string = "34F0A37AAD20F4A260F0A5B3CB3D7FB50673212263E58A380BC10474BB039CE4"

var TestGenesisBlock = FromJson([]byte(`{
	"type": "open",
	"source": "B0311EA55708D6A53C75CDBF88300259C6D018522FE3D4D0A242E431F9E8B6D0",
	"representative": "nano_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
	"account": "nano_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
	"work": "9680625b39d3363d",
	"signature": "ECDA914373A2F0CA1296475BAEE40500A7F0A7AD72A5A80C81D7FAB7F6C802B2CC7DB50F5DD0FB25B2EF11761FA7344A158DD5A700B21BD47DE5BD0F63153A02"
}`)).(*OpenBlock)

var LiveGenesisBlock = FromJson([]byte(`{
	"type":           "open",
	"source":         "E89208DD038FBB269987689621D52292AE9C35941A7484756ECCED92A65093BA",
	"representative": "nano_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3",
	"account":        "nano_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3",
	"work":           "62f05417dd3fb691",
	"signature":      "9F0C933C8ADE004D808EA1985FA746A7E95BA2A38F867640F53EC8F180BDFE9E2C1268DEAD7C2664F356E37ABA362BC58E46DBA03E523A7B5A19E4B6EB12BB02"
}`)).(*OpenBlock)

type BlockType string

const (
	Open    BlockType = "open"
	Receive           = "receive"
	Send              = "send"
	Change            = "change"
)

type Block interface {
	Type() BlockType
	GetSignature() types.Signature
	GetWork() types.Work
	RootHash() types.BlockHash
	Hash() types.BlockHash
	PreviousBlockHash() types.BlockHash
}

type CommonBlock struct {
	Work      types.Work
	Signature types.Signature
	Confirmed bool
}

type OpenBlock struct {
	SourceHash     types.BlockHash
	Representative types.Account
	Account        types.Account
	CommonBlock
}

type SendBlock struct {
	PreviousHash types.BlockHash
	Destination  types.Account
	Balance      uint128.Uint128
	CommonBlock
}

type ReceiveBlock struct {
	PreviousHash types.BlockHash
	SourceHash   types.BlockHash
	CommonBlock
}

type ChangeBlock struct {
	PreviousHash   types.BlockHash
	Representative types.Account
	CommonBlock
}

func (b *OpenBlock) Hash() types.BlockHash {
	return types.BlockHashFromBytes(HashOpen(b.SourceHash, b.Representative, b.Account))
}

func (b *ReceiveBlock) Hash() types.BlockHash {
	return types.BlockHashFromBytes(HashReceive(b.PreviousHash, b.SourceHash))
}

func (b *ChangeBlock) Hash() types.BlockHash {
	return types.BlockHashFromBytes(HashChange(b.PreviousHash, b.Representative))
}

func (b *SendBlock) Hash() types.BlockHash {
	return types.BlockHashFromBytes(HashSend(b.PreviousHash, b.Destination, b.Balance))
}

func (b *ReceiveBlock) PreviousBlockHash() types.BlockHash {
	return b.PreviousHash
}

func (b *ChangeBlock) PreviousBlockHash() types.BlockHash {
	return b.PreviousHash
}

func (b *SendBlock) PreviousBlockHash() types.BlockHash {
	return b.PreviousHash
}

func (b *OpenBlock) PreviousBlockHash() types.BlockHash {
	return b.SourceHash
}

func (b *OpenBlock) RootHash() types.BlockHash {
	pub, _ := address.AddressToPub(b.Account)
	return types.BlockHash(hex.EncodeToString(pub))
}

func (b *ReceiveBlock) RootHash() types.BlockHash {
	return b.PreviousHash
}

func (b *ChangeBlock) RootHash() types.BlockHash {
	return b.PreviousHash
}

func (b *SendBlock) RootHash() types.BlockHash {
	return b.PreviousHash
}

func (b *CommonBlock) GetSignature() types.Signature {
	return b.Signature
}

func (b *CommonBlock) GetWork() types.Work {
	return b.Work
}

func (*SendBlock) Type() BlockType {
	return Send
}

func (*OpenBlock) Type() BlockType {
	return Open
}

func (*ChangeBlock) Type() BlockType {
	return Change
}

func (*ReceiveBlock) Type() BlockType {
	return Receive
}

func (b *OpenBlock) VerifySignature() (bool, error) {
	pub, _ := address.AddressToPub(b.Account)
	res := ed25519.Verify(pub, b.Hash().ToBytes(), b.Signature.ToBytes())
	return res, nil
}

type RawBlock struct {
	Type           BlockType
	Source         types.BlockHash
	Representative types.Account
	Account        types.Account
	Work           types.Work
	Signature      types.Signature
	Previous       types.BlockHash
	Balance        uint128.Uint128
	Destination    types.Account
}

func FromJson(b []byte) (block Block) {
	var raw RawBlock
	json.Unmarshal(b, &raw)
	common := CommonBlock{
		Work:      raw.Work,
		Signature: raw.Signature,
	}

	switch raw.Type {
	case Open:
		b := OpenBlock{
			raw.Source,
			raw.Representative,
			raw.Account,
			common,
		}
		block = &b
	case Send:
		b := SendBlock{
			raw.Previous,
			raw.Destination,
			raw.Balance,
			common,
		}
		block = &b
	case Receive:
		b := ReceiveBlock{
			raw.Previous,
			raw.Source,
			common,
		}
		block = &b
	case Change:
		b := ChangeBlock{
			raw.Previous,
			raw.Representative,
			common,
		}
		block = &b
	default:
		panic("Unknown block type")
	}

	return block

}

func (b RawBlock) Hash() (result []byte) {
	switch b.Type {
	case Open:
		return HashOpen(b.Source, b.Representative, b.Account)
	case Send:
		return HashSend(b.Previous, b.Destination, b.Balance)
	case Receive:
		return HashReceive(b.Previous, b.Source)
	case Change:
		return HashChange(b.Previous, b.Representative)
	default:
		panic("Unknown block type! " + b.Type)
	}
}

func (b RawBlock) HashToString() (result types.BlockHash) {
	return types.BlockHash(strings.ToUpper(hex.EncodeToString(b.Hash())))
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

func HashReceive(previous types.BlockHash, source types.BlockHash) (result []byte) {
	previous_bytes, _ := hex.DecodeString(string(previous))
	source_bytes, _ := hex.DecodeString(string(source))
	return HashBytes(previous_bytes, source_bytes)
}

func HashChange(previous types.BlockHash, representative types.Account) (result []byte) {
	previous_bytes, _ := hex.DecodeString(string(previous))
	repr_bytes, _ := address.AddressToPub(representative)
	return HashBytes(previous_bytes, repr_bytes)
}

func HashSend(previous types.BlockHash, destination types.Account, balance uint128.Uint128) (result []byte) {
	previous_bytes, _ := hex.DecodeString(string(previous))
	dest_bytes, _ := address.AddressToPub(destination)
	balance_bytes := balance.GetBytes()

	return HashBytes(previous_bytes, dest_bytes, balance_bytes)
}

func HashOpen(source types.BlockHash, representative types.Account, account types.Account) (result []byte) {
	source_bytes, _ := hex.DecodeString(string(source))
	repr_bytes, _ := address.AddressToPub(representative)
	account_bytes, _ := address.AddressToPub(account)
	return HashBytes(source_bytes, repr_bytes, account_bytes)
}

// ValidateWork takes the "work" value (little endian from hex)
// and block hash and verifies that the work passes the difficulty.
// To verify this, we create a new 8 byte hash of the
// work and the block hash and convert this to a uint64
// which must be higher (or equal) than the difficulty
// (0xffffffc000000000) to be valid.
func ValidateWork(block_hash []byte, work []byte) bool {
	hash, err := blake2b.New(8, nil)
	if err != nil {
		panic("Unable to create hash")
	}
	return validateWork(hash, block_hash, work)
}

func validateWork(digest hash.Hash, block []byte, work []byte) bool {
	digest.Reset()
	digest.Write(work)
	digest.Write(block)

	sum := digest.Sum(nil)
	return binary.LittleEndian.Uint64(sum) >= WorkThreshold
}

func validateNonce(digest hash.Hash, block []byte, nonce uint64) bool {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, nonce)
	return validateWork(digest, block, b)
}

func ValidateBlockWork(b Block) bool {
	hash_bytes := b.RootHash().ToBytes()
	work_bytes, _ := hex.DecodeString(string(b.GetWork()))

	res := ValidateWork(hash_bytes, utils.Reversed(work_bytes))
	return res
}

func GenerateWorkForHash(b types.BlockHash) types.Work {
	block_hash := b.ToBytes()
	digest, err := blake2b.New(8, nil)
	if err != nil {
		panic("Unable to create hash")
	}
	var nonce uint64
	for ; !validateNonce(digest, block_hash, nonce); nonce++ {
	}
	work := make([]byte, 8)
	binary.BigEndian.PutUint64(work, nonce)
	return types.Work(hex.EncodeToString(work))
}

func GenerateWork(b Block) types.Work {
	return GenerateWorkForHash(b.Hash())
}
