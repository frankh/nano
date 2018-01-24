package blocks

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/frankh/rai/address"
	"github.com/frankh/rai/utils"
)

var TestConfigLive = Config{
	":memory:",
	LiveGenesisBlock,
	0xffffffc000000000,
}

var TestConfigTest = Config{
	":memory:",
	TestGenesisBlock,
	0xff00000000000000,
}

func TestSignMessage(t *testing.T) {
	privateKeyData := "34F0A37AAD20F4A260F0A5B3CB3D7FB50673212263E58A380BC10474BB039CE4"

	block := FromJson([]byte(`{
		"type":           "open",
		"source":         "B0311EA55708D6A53C75CDBF88300259C6D018522FE3D4D0A242E431F9E8B6D0",
		"representative": "xrb_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
		"account":        "xrb_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
		"work":           "9680625b39d3363d",
		"signature":      "ECDA914373A2F0CA1296475BAEE40500A7F0A7AD72A5A80C81D7FAB7F6C802B2CC7DB50F5DD0FB25B2EF11761FA7344A158DD5A700B21BD47DE5BD0F63153A02"
	}`))

	signatureBytes := SignMessage(privateKeyData, block.Hash().ToBytes())
	signature := strings.ToUpper(hex.EncodeToString(signatureBytes))

	if signature != string(block.GetSignature()) {
		t.Errorf("Signature %s was expected to be %s", signature, block.GetSignature())
	}

}

func TestGenerateWork(t *testing.T) {
	Init(TestConfigTest)
	GenerateWork(LiveGenesisBlock)
}

func BenchmarkGenerateWork(b *testing.B) {
	Init(Config{
		":memory:",
		LiveGenesisBlock,
		0xfff0000000000000,
	})
	for n := 0; n < b.N; n++ {
		GenerateWork(LiveGenesisBlock)
	}
}

func TestValidateWork(t *testing.T) {
	Init(TestConfigLive)

	liveBlockHash, _ := address.AddressToPub(LiveGenesisBlock.Account)
	liveWorkBytes, _ := hex.DecodeString(string(LiveGenesisBlock.Work))
	liveBadWork, _ := hex.DecodeString("0000000000000000")

	if !ValidateBlockWork(LiveGenesisBlock) {
		t.Errorf("Work validation failed for genesis block")
		return
	}

	// A bit of a redundandy test to ensure ValidateBlockWork is correct
	if !ValidateWork(liveBlockHash, utils.Reversed(liveWorkBytes)) {
		t.Errorf("Work validation failed for genesis block")
	}

	if ValidateWork(liveBlockHash, liveBadWork) {
		t.Errorf("Work validation passed for bad work")
	}
}

func TestHashOpen(t *testing.T) {
	if LiveGenesisBlock.Hash() != LiveGenesisBlockHash {
		t.Errorf("Genesis block hash is not correct, expected %s, got %s", LiveGenesisBlockHash, LiveGenesisBlock.Hash())
	}
}
