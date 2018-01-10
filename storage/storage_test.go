package storage

import (
	"github.com/frankh/rai/blocks"
	"testing"
)

func TestInit(t *testing.T) {
	Init(":memory:")

}

func TestGenesisBalance(t *testing.T) {
	Init(":memory:")

	block := FetchBlock(blocks.LiveGenesisBlockHash)

	if block.GetBalance().String() != "ffffffffffffffffffffffffffffffff" {
		t.Errorf("Genesis block has invalid initial balance")
	}
}
