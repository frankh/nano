package blocks

import (
	"testing"
)

func TestInit(t *testing.T) {
	Init(TestConfigLive)

}

func TestGenesisBalance(t *testing.T) {
	Init(TestConfigLive)

	block := FetchBlock(LiveGenesisBlockHash)

	if block.GetBalance().String() != "ffffffffffffffffffffffffffffffff" {
		t.Errorf("Genesis block has invalid initial balance")
	}
}

func TestMissingBlock(t *testing.T) {
	Init(TestConfigTest)

	block := FetchBlock(LiveGenesisBlockHash)

	if block != nil {
		t.Errorf("Found live genesis on test config")
	}
}
