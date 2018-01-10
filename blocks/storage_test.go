package blocks

import (
	"testing"
)

func TestInit(t *testing.T) {
	Init(":memory:")

}

func TestGenesisBalance(t *testing.T) {
	Init(":memory:")

	block := FetchBlock(LiveGenesisBlockHash)

	if block.GetBalance().String() != "ffffffffffffffffffffffffffffffff" {
		t.Errorf("Genesis block has invalid initial balance")
	}
}
