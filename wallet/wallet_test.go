package wallet

import (
	"github.com/frankh/rai/blocks"
	"testing"
)

var TestConfigTest = blocks.Config{
	":memory:",
	blocks.TestGenesisBlock,
}

func TestNew(t *testing.T) {
	blocks.Init(TestConfigTest)

	w := New(blocks.TestPrivateKey)
	if w.GetBalance() != blocks.GenesisAmount {
		t.Errorf("Genesis block doesn't have correct balance")
	}
}
