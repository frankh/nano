package blocks

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/frankh/nano/address"
	"github.com/frankh/nano/utils"
)

func TestSignMessage(t *testing.T) {
	test_private_key_data := "34F0A37AAD20F4A260F0A5B3CB3D7FB50673212263E58A380BC10474BB039CE4"

	block := FromJson([]byte(`{
		"type":           "open",
		"source":         "B0311EA55708D6A53C75CDBF88300259C6D018522FE3D4D0A242E431F9E8B6D0",
		"representative": "nano_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
		"account":        "nano_3e3j5tkog48pnny9dmfzj1r16pg8t1e76dz5tmac6iq689wyjfpiij4txtdo",
		"work":           "9680625b39d3363d",
		"signature":      "ECDA914373A2F0CA1296475BAEE40500A7F0A7AD72A5A80C81D7FAB7F6C802B2CC7DB50F5DD0FB25B2EF11761FA7344A158DD5A700B21BD47DE5BD0F63153A02"
	}`))

	signature_bytes := SignMessage(test_private_key_data, block.Hash().ToBytes())
	signature := strings.ToUpper(hex.EncodeToString(signature_bytes))

	if signature != string(block.GetSignature()) {
		t.Errorf("Signature %s was expected to be %s", signature, block.GetSignature())
	}

}

func TestGenerateWork(t *testing.T) {
	WorkThreshold = 0xfff0000000000000
	GenerateWork(LiveGenesisBlock)
}

func BenchmarkGenerateWork(b *testing.B) {
	WorkThreshold = 0xfff0000000000000
	for n := 0; n < b.N; n++ {
		GenerateWork(LiveGenesisBlock)
	}
}

func TestValidateWork(t *testing.T) {
	WorkThreshold = 0xffffffc000000000

	live_block_hash, _ := address.AddressToPub(LiveGenesisBlock.Account)
	live_work_bytes, _ := hex.DecodeString(string(LiveGenesisBlock.Work))
	live_bad_work, _ := hex.DecodeString("0000000000000000")

	if !ValidateBlockWork(LiveGenesisBlock) {
		t.Errorf("Work validation failed for genesis block")
		return
	}

	// A bit of a redundandy test to ensure ValidateBlockWork is correct
	if !ValidateWork(live_block_hash, utils.Reversed(live_work_bytes)) {
		t.Errorf("Work validation failed for genesis block")
	}

	if ValidateWork(live_block_hash, live_bad_work) {
		t.Errorf("Work validation passed for bad work")
	}

}

func TestHashOpen(t *testing.T) {
	if LiveGenesisBlock.Hash() != LiveGenesisBlockHash {
		t.Errorf("Genesis block hash is not correct, expected %s, got %s", LiveGenesisBlockHash, LiveGenesisBlock.Hash())
	}
}
