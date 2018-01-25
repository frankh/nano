package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/frankh/rai/address"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
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

func TestPoW(t *testing.T) {
	blocks.Init(TestConfigTest)
	w := New(blocks.TestPrivateKey)

	if w.GeneratePoWAsync() != nil || !w.WaitingForPoW() {
		t.Errorf("Failed to start PoW generation")
	}

	if w.GeneratePoWAsync() == nil {
		t.Errorf("Started PoW while already in progress")
	}

	_, err := w.Send(blocks.TestGenesisBlock.Account, uint128.FromInts(0, 1))

	if err == nil {
		t.Errorf("Created send block without PoW")
	}

	w.WaitPoW()

	send, _ := w.Send(blocks.TestGenesisBlock.Account, uint128.FromInts(0, 1))

	if !blocks.ValidateBlockWork(send) {
		t.Errorf("Invalid work")
	}

}

func TestSend(t *testing.T) {
	blocks.Init(TestConfigTest)
	w := New(blocks.TestPrivateKey)

	w.GeneratePowSync()
	amount := uint128.FromInts(1, 1)

	sendBlock, _ := w.Send(blocks.TestGenesisBlock.Account, amount)

	if w.GetBalance() != blocks.GenesisAmount.Sub(amount) {
		t.Errorf("Balance unchanged after send")
	}

	_, err := w.Send(blocks.TestGenesisBlock.Account, blocks.GenesisAmount)
	if err == nil {
		t.Errorf("Sent more than account balance")
	}

	w.GeneratePowSync()
	blocks.StoreBlock(sendBlock)
	w.Receive(sendBlock.Hash())

	if w.GetBalance() != blocks.GenesisAmount {
		t.Errorf("Balance not updated after receive")
	}
}

func TestOpen(t *testing.T) {
	blocks.Init(TestConfigTest)
	amount := uint128.FromInts(1, 1)

	sendWallet := New(blocks.TestPrivateKey)
	sendWallet.GeneratePowSync()

	_, privKey := address.GenerateKey()
	openWallet := New(hex.EncodeToString(privKey))
	sendBlock, _ := sendWallet.Send(openWallet.Address(), amount)
	openWallet.GeneratePowSync()

	_, err := openWallet.Open(sendBlock.Hash(), openWallet.Address())
	if err == nil {
		t.Errorf("Expected error for referencing unstored send")
	}

	if openWallet.GetBalance() != uint128.FromInts(0, 0) {
		t.Errorf("Open should start at zero balance")
	}

	blocks.StoreBlock(sendBlock)
	_, err = openWallet.Open(sendBlock.Hash(), openWallet.Address())
	if err != nil {
		t.Errorf("Open block failed: %s", err)
	}

	if openWallet.GetBalance() != amount {
		t.Errorf("Open balance didn't equal send amount")
	}

	_, err = openWallet.Open(sendBlock.Hash(), openWallet.Address())
	if err == nil {
		t.Errorf("Expected error for creating duplicate open block")
	}
}
