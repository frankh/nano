package wallet

import (
	"github.com/frankh/rai/blocks"
	"testing"
)

var TestConfigTest = blocks.Config{
	":memory:",
	blocks.TestGenesisBlock,
	0xfffffff000000000,
}
var TestConfigLive = blocks.Config{
	":memory:",
	blocks.LiveGenesisBlock,
	0xffffff0000000000,
}

func TestNew(t *testing.T) {
	blocks.Init(TestConfigTest)

	w := New(blocks.TestPrivateKey)
	if w.GetBalance() != blocks.GenesisAmount {
		t.Errorf("Genesis block doesn't have correct balance")
	}
}

func TestPoWFail(t *testing.T) {
	blocks.Init(TestConfigLive)
	w := New(blocks.TestPrivateKey)
	err := w.GeneratePoWAsync()
	if err == nil {
		t.Errorf("Empty wallet should not generate work %#v", err)
	}
}

func TestPoW(t *testing.T) {
	blocks.Init(TestConfigTest)
	w := New(blocks.TestPrivateKey)

	if w.GeneratePoWAsync() != nil {
		t.Errorf("Failed to start PoW generation")
	}

	w.WaitPoW()
}
