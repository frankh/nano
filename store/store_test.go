package store

import (
	"testing"

	"github.com/frankh/rai/blocks"
)

func TestInit(t *testing.T) {
	Init(TestConfigLive)

}

func TestGenesisBalance(t *testing.T) {
	Init(TestConfigLive)

	block := FetchBlock(blocks.LiveGenesisBlockHash)

	if GetBalance(block).String() != "ffffffffffffffffffffffffffffffff" {
		t.Errorf("Genesis block has invalid initial balance")
	}
}

func TestMissingBlock(t *testing.T) {
	Init(TestConfig)

	block := FetchBlock(blocks.LiveGenesisBlockHash)

	if block != nil {
		t.Errorf("Found live genesis on test config")
	}
}
