package wallet

import (
	"github.com/frankh/rai/blocks"
	"testing"
)

var TestConfigTest = blocks.Config{
	":memory:",
	blocks.TestGenesisBlock,
	0xff00000000000000,
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

	if w.WaitingForPoW() {
		t.Errorf("Marked as generating when not")
	}
}

func TestPoW(t *testing.T) {
	blocks.Init(TestConfigTest)
	w := New(blocks.TestPrivateKey)

	if w.GeneratePoWAsync() != nil || !w.WaitingForPoW() {
		t.Errorf("Failed to start PoW generation")
	}

	if w.GeneratePoWAsync() == nil {
		t.Errorf("Started PoW while already in progress")
	}

	w.WaitPoW()
	w.Head.(*blocks.OpenBlock).Work = *w.Work

	if !blocks.ValidateBlockWork(w.Head) {
		t.Errorf("Pow was invalid")
	}

}
