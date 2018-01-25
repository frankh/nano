package blocks

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
const LiveGenesisSourceHash rai.BlockHash = "E89208DD038FBB269987689621D52292AE9C35941A7484756ECCED92A65093BA"

var GenesisAmount uint128.Uint128 = uint128.FromInts(0xffffffffffffffff, 0xffffffffffffffff)

const TestPrivateKey string = "34F0A37AAD20F4A260F0A5B3CB3D7FB50673212263E58A380BC10474BB039CE4"

var TestGenesisBlock = FromJson([]byte(`{
	"type": "open",
	"source": "B0311EA55708D6A53C75CDBF88300259C6D018522FE3D4D0A242E431F9E8B6D0",
	"representative": "xrb_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
	"account": "xrb_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
	"work": "9680625b39d3363d",
	"signature": "ECDA914373A2F0CA1296475BAEE40500A7F0A7AD72A5A80C81D7FAB7F6C802B2CC7DB50F5DD0FB25B2EF11761FA7344A158DD5A700B21BD47DE5BD0F63153A02"
}`)).(*OpenBlock)

var LiveGenesisBlock = FromJson([]byte(`{
	"type":           "open",
	"source":         "E89208DD038FBB269987689621D52292AE9C35941A7484756ECCED92A65093BA",
	"representative": "xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3",
	"account":        "xrb_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3",
	"work":           "62f05417dd3fb691",
	"signature":      "9F0C933C8ADE004D808EA1985FA746A7E95BA2A38F867640F53EC8F180BDFE9E2C1268DEAD7C2664F356E37ABA362BC58E46DBA03E523A7B5A19E4B6EB12BB02"
}`)).(*OpenBlock)

var LiveConfig = Config{
	"db.sqlite",
	LiveGenesisBlock,
	0xffffffc000000000,
}
var TestConfig = Config{
	":memory:",
	TestGenesisBlock,
	0xffffffc000000000,
}

type BlockType string

const (
	Open    BlockType = "open"
	Receive           = "receive"
	Send              = "send"
	Change            = "change"
)

type Block interface {
	Type() BlockType
	GetBalance() uint128.Uint128
	GetSignature() rai.Signature
	GetWork() rai.Work
	RootHash() rai.BlockHash
	Hash() rai.BlockHash
	Previous() Block
}

type CommonBlock struct {
	Work      rai.Work
	Signature rai.Signature
	Connected bool
	Confirmed bool
}

type OpenBlock struct {
	SourceHash     rai.BlockHash
	Representative rai.Account
	Account        rai.Account
	CommonBlock
}

type SendBlock struct {
	PreviousHash rai.BlockHash
	Destination  rai.Account
	Balance      uint128.Uint128
	CommonBlock
}

type ReceiveBlock struct {
	PreviousHash rai.BlockHash
	SourceHash   rai.BlockHash
	CommonBlock
}

type ChangeBlock struct {
	PreviousHash   rai.BlockHash
	Representative rai.Account
	CommonBlock
}

func (b *OpenBlock) Hash() rai.BlockHash {
	return rai.BlockHashFromBytes(HashOpen(b.SourceHash, b.Representative, b.Account))
}

func (b *ReceiveBlock) Hash() rai.BlockHash {
	return rai.BlockHashFromBytes(HashReceive(b.PreviousHash, b.SourceHash))
}

func (b *ChangeBlock) Hash() rai.BlockHash {
	return rai.BlockHashFromBytes(HashChange(b.PreviousHash, b.Representative))
}

func (b *SendBlock) Hash() rai.BlockHash {
	return rai.BlockHashFromBytes(HashSend(b.PreviousHash, b.Destination, b.Balance))
}

func (b *OpenBlock) Source() *SendBlock {
	return FetchBlock(b.SourceHash).(*SendBlock)
}

func (b *ReceiveBlock) Source() *SendBlock {
	return FetchBlock(b.SourceHash).(*SendBlock)
}

func (b *ReceiveBlock) Previous() Block {
	return FetchBlock(b.PreviousHash)
}

func (b *ChangeBlock) Previous() Block {
	return FetchBlock(b.PreviousHash)
}

func (b *SendBlock) Previous() Block {
	return FetchBlock(b.PreviousHash)
}

func (b *OpenBlock) Previous() Block {
	return FetchBlock(b.SourceHash)
}

func (b *OpenBlock) RootHash() rai.BlockHash {
	pub, _ := address.AddressToPub(b.Account)
	return rai.BlockHash(hex.EncodeToString(pub))
}

func (b *ReceiveBlock) RootHash() rai.BlockHash {
	return b.PreviousHash
}

func (b *ChangeBlock) RootHash() rai.BlockHash {
	return b.PreviousHash
}

func (b *SendBlock) RootHash() rai.BlockHash {
	return b.PreviousHash
}

func (b *CommonBlock) GetSignature() rai.Signature {
	return b.Signature
}

func (b *CommonBlock) GetWork() rai.Work {
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
	Source         rai.BlockHash
	Representative rai.Account
	Account        rai.Account
	Work           rai.Work
	Signature      rai.Signature
	Previous       rai.BlockHash
	Balance        uint128.Uint128
	Destination    rai.Account
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

func (b RawBlock) HashToString() (result rai.BlockHash) {
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
	if len(work) != 8 {
		panic("Bad work length")
	}

	hash.Write(work)
	hash.Write(block_hash)

	work_value := hash.Sum(nil)
	work_value_int := binary.LittleEndian.Uint64(work_value)

	return work_value_int >= Conf.WorkThreshold
}

func ValidateBlockWork(b Block) bool {
	hash_bytes := b.RootHash().ToBytes()
	work_bytes, _ := hex.DecodeString(string(b.GetWork()))

	res := ValidateWork(hash_bytes, utils.Reversed(work_bytes))
	return res
}

func GenerateWorkForHash(b rai.BlockHash) rai.Work {
	block_hash := b.ToBytes()
	work := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	for {
		if ValidateWork(block_hash, work) {
			return rai.Work(fmt.Sprintf("%x", utils.Reversed(work)))
		}
		incrementWork(work)
	}
}

func GenerateWork(b Block) rai.Work {
	return GenerateWorkForHash(b.Hash())
}

func incrementWork(work []byte) {
	for i := 0; i < len(work)-1; i++ {
		if work[i] < 255 {
			work[i]++
			return
		} else {
			work[i]++
			incrementWork(work[i+1:])
			return
		}
	}
}
