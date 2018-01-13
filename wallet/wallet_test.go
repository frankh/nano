package wallet

import (
	"encoding/hex"
	"github.com/frankh/rai/address"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
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

	send, _ := w.Send(blocks.TestGenesisBlock.Account, amount)

	if w.GetBalance() != blocks.GenesisAmount.Sub(amount) {
		t.Errorf("Balance unchanged after send")
	}

	_, err := w.Send(blocks.TestGenesisBlock.Account, blocks.GenesisAmount)
	if err == nil {
		t.Errorf("Sent more than account balance")
	}

	w.GeneratePowSync()
	blocks.StoreBlock(send)
	w.Receive(send.Hash())

	if w.GetBalance() != blocks.GenesisAmount {
		t.Errorf("Balance not updated after receive")
	}

}

func TestOpen(t *testing.T) {
	blocks.Init(TestConfigTest)
	amount := uint128.FromInts(1, 1)

	blocks.Init(TestConfigTest)
	sendW := New(blocks.TestPrivateKey)
	_, priv := address.GenerateKey()
	openW := New(hex.EncodeToString(priv))

	sendW.GeneratePowSync()
	send, _ := sendW.Send(openW.Address(), amount)

	openWork := blocks.GenerateWork(send)
	_, err := openW.Open(send.Hash(), openW.Address(), &openWork)

	if err == nil {
		t.Errorf("Expected error for referencing unstored send")
	}

	if openW.GetBalance() != uint128.FromInts(0, 0) {
		t.Errorf("Open should start at zero balance")
	}

	blocks.StoreBlock(send)
	_, err = openW.Open(send.Hash(), openW.Address(), &openWork)
	if err != nil {
		t.Errorf("Open block failed")
	}

	if openW.GetBalance() != amount {
		t.Errorf("Open balance didn't equal send amount")
	}

}
