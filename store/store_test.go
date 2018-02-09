package store

import (
	"os"
	"testing"

	"github.com/frankh/nano/blocks"
)

func TestInit(t *testing.T) {
	Init(TestConfigLive)

	os.RemoveAll(TestConfigLive.Path)
}

func TestGenesisBalance(t *testing.T) {
	Init(TestConfigLive)

	block := FetchBlock(blocks.LiveGenesisBlockHash)

	if GetBalance(block).String() != "ffffffffffffffffffffffffffffffff" {
		t.Errorf("Genesis block has invalid initial balance")
	}
	os.RemoveAll(TestConfigLive.Path)
}

func TestMissingBlock(t *testing.T) {
	Init(TestConfig)

	block := FetchBlock(blocks.LiveGenesisBlockHash)

	if block != nil {
		t.Errorf("Found live genesis on test config")
	}
	os.RemoveAll(TestConfig.Path)
}
