package node

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/frankh/crypto/ed25519"
	"github.com/frankh/nano/blocks"
	"github.com/frankh/nano/store"
	"github.com/frankh/nano/types"
)

var publishSend, _ = hex.DecodeString("5243050501030002B6460102018F076CC32FF2F65AD397299C47F8CA2BE784D5DE394D592C22BE8BFFBE91872F1D2A2BCC1CB47FB854D6D31E43C6391EADD5750BB9689E5DF0D6CB0000003D11C83DBCFF748EB4B7F7A3C059DDEEE5C8ECCC8F20DEF3AF3C4F0726F879082ED051D0C62A54CD69C4A66B020369B7033C5B0F77654173AB24D5C7A64CC4FFF0BDB368FCC989E41A656569047627C49A2A6D2FBC")
var publishReceive, _ = hex.DecodeString("5243050501030003233FF43F2ADE055D4D4BCC1C19A3100B720C21E5548A547B9B21938BBDBB19EE28A1763099135DADB3F223C0A4138269C7146A6431AF0597D24276BB0A24BAFCBA254A264BAA0BCBA5962A77E15D4EB021043FFFEA9E4391E179D467C66C69675E9634F9C124060FC65D5B2F67FCA38E8BA93BF910EB337010BC51E652B0640D62F2642CB37BCD7C")
var publishTest, _ = hex.DecodeString("52430505010300030AFC4456F1A54722B101E41B1C2E3F7AF0EFD456EAE3621786C021D72C0BA9880FD491C3FF52227C8CDF76C88CE8F650320042349210AD2681134FD74080675C60734FAA7F89DDF5BDA156A5C7996A79F2CBD22E244B4E39D497261D356A30BE70973313A71A7D52700A560191B8A926FCE44B987A96FE61A8C469BBE383340831783CA6A6511D6A")
var publishOpen, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC005A8B4F60DDA9D288B1C14A4CB")
var publishChange, _ = hex.DecodeString("5243050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED35BBAD0F1DE013836D3471C1")
var publishWrongBlock, _ = hex.DecodeString("5243050501030002611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED35BBAD0F1DE013836D3471C1")
var publishWrongMagic, _ = hex.DecodeString("5242050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED35BBAD0F1DE013836D3471C1")
var publishWrongSig, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC015A8B4F60DDA9D288B1C14A4CB")
var publishWrongWork, _ = hex.DecodeString("5242050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED34BBAD0F1DE013836D3471C0")
var keepAlive, _ = hex.DecodeString("524305050102000000000000000000000000FFFF49B13E26A31B00000000000000000000FFFF637887DF340400000000000000000000FFFFCC2C6D15A31B00000000000000000000FFFF5EC16857239C00000000000000000000FFFF23BD2D1FA31B00000000000000000000FFFF253B710AA31B00000000000000000000FFFF50740256A7E500000000000000000000FFFF4631D644A31B")

var confirmAck, _ = hex.DecodeString("524305050105000289aaf8e5f19f60ebc9476f382dbee256deae2695b47934700d9aad49d86ccb249ceb5c2840fe3fdf2dcb9c40e142181e7bd158d07ca3f8388dc3b3c0acd395d85b38e04ce1dac45b070957046d31eb7f58caaa777a5e13d85fe2aae7514b490e9c1dd00100000000aef053ab1832d41df356290a704e6c6c47787c6da4710ee2399e60e0ab607e9e51380a2c22710ed4018392474228b4e7c80f1c6714dcc3c9ef4befa563ecc35905bd9a62bd5b7ebdc5ebc9f576392e00445a07742dc4b2bc1355aef245522b19ae5640985f7759954ebf5147a125fec7e9f1973cf1d2a9d182c9223392b4cc10cdb11bca27c455ec8b13f4482b506d02576cfad0046c5f1c")
var confirmReq, _ = hex.DecodeString("52430505010400030c32f8cac423ec13236e09db435a18471ef39274959e6f8b44f005577614190e6e470adf874730bb15f067e04ec4ccd77426e69166a72d57d592a4e15eff1df97560262045e5a612c015205a5e73a53fe3775bd5809f6723641b31c7b103ebb30adc93932c7fba8c0a29d8ca1fb22514a2490552dcdb028401975cd8c9014b0fccd88343ef983eae")

func TestReadWriteMessageKeepAlive(t *testing.T) {
	var message MessageKeepAlive
	buf := bytes.NewBuffer(keepAlive)
	err := message.Read(buf)
	if err != nil {
		t.Errorf("Failed to read keepalive: %s", err)
	}
	if len(message.Peers) != 8 {
		t.Errorf("Wrong number of keepalive peers %d", len(message.Peers))
	}

	buf = bytes.NewBuffer(publishChange)
	if message.Read(buf) == nil {
		t.Errorf("Should fail to read wrong message type")
	}

	if message.Peers[0].Port != 7075 {
		t.Errorf("Wrong port deserialization")
	}

	if message.Peers[0].IP.String() != "73.177.62.38" {
		t.Errorf("Wrong IP deserialization")
	}

	var writeBuf bytes.Buffer
	err = message.Write(&writeBuf)
	if err != nil {
		t.Errorf("Failed to write keepalive: %s", err)
	}

	if !bytes.Equal(keepAlive, writeBuf.Bytes()) {
		t.Errorf("Failed to rewrite keepalive message\n%x\n%x\n", keepAlive, writeBuf.Bytes())
	}
}

func TestReadWriteConfirmAck(t *testing.T) {
	var m MessageConfirmAck
	buf := bytes.NewBuffer(confirmAck)
	err := m.Read(buf)
	if err != nil {
		t.Errorf("Failed to read message")
	}

	var writeBuf bytes.Buffer
	err = m.Write(&writeBuf)
	if err != nil {
		t.Errorf("Failed to write message")
	}

	if bytes.Compare(confirmAck, writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly")
	}

	block := m.ToBlock().(*blocks.SendBlock)
	if !blocks.ValidateBlockWork(block) {
		t.Errorf("Work validation failed")
	}

	publicKey := ed25519.PublicKey(m.Account[:])
	if !ed25519.Verify(publicKey, m.MessageVote.Hash(), m.Signature[:]) {
		t.Errorf("Failed to validate signature")
	}
}

func TestReadWriteConfirmReq(t *testing.T) {
	var m MessageConfirmReq
	buf := bytes.NewBuffer(confirmReq)
	err := m.Read(buf)
	if err != nil {
		t.Errorf("Failed to read message")
	}

	var writeBuf bytes.Buffer
	err = m.Write(&writeBuf)
	if err != nil {
		t.Errorf("Failed to write message")
	}

	if bytes.Compare(confirmReq, writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly")
	}

	block := m.ToBlock().(*blocks.ReceiveBlock)
	if !blocks.ValidateBlockWork(block) {
		t.Errorf("Work validation failed")
	}
}

func TestReadWriteMessagePublish(t *testing.T) {
	var m MessagePublish
	buf := bytes.NewBuffer(publishOpen)
	err := m.Read(buf)
	if err != nil {
		t.Errorf("Failed to read message")
	}

	var writeBuf bytes.Buffer
	err = m.Write(&writeBuf)
	if err != nil {
		t.Errorf("Failed to write message")
	}

	if bytes.Compare(publishOpen, writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly")
	}

	block := m.ToBlock().(*blocks.OpenBlock)
	if !blocks.ValidateBlockWork(block) {
		t.Errorf("Work validation failed")
	}

	if block.Account != "nano_14jyjetsh8p7jxx1of38ctsa779okt9d1pdnmtjpqiukuq8zugr3bxpxf1zu" {
		t.Errorf("Deserialised account badly")
	}
}

func validateTestBlock(t *testing.T, b blocks.Block, expectedHash types.BlockHash) {
	if b.Hash() != expectedHash {
		t.Errorf("Wrong blockhash %s", b.Hash())
	}
	if !blocks.ValidateBlockWork(b) {
		t.Errorf("Bad PoW")
	}
	if b.Type() == blocks.Open {
		passed, _ := b.(*blocks.OpenBlock).VerifySignature()
		if !passed {
			t.Errorf("Failed to verify signature")
		}
	}
}

func TestReadPublish(t *testing.T) {
	var m MessagePublish
	err := m.Read(bytes.NewBuffer(publishSend))
	if err != nil {
		t.Errorf("Failed to read send message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), types.BlockHash("687DCB9C8EB8AF9F39D8107C3432A8732EDBED1E3B5E2E0F6B86643D1EB5E24F"))

	err = m.Read(bytes.NewBuffer(publishReceive))
	if err != nil {
		t.Errorf("Failed to read receive message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), types.BlockHash("7D3E9D79342AA73B7148CB46706D23ED8BB0041A5316D67A053F336ABF0E6B60"))

	err = m.Read(bytes.NewBuffer(publishOpen))
	if err != nil {
		t.Errorf("Failed to read open message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), types.BlockHash("5F73CF212E58563734D57CCFCCEFE481DE40C96F097F594F4FA32C5585D84AA4"))

	err = m.Read(bytes.NewBuffer(publishChange))
	if err != nil {
		t.Errorf("Failed to read change message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), types.BlockHash("4AABA9923AC794B635B8C3CC275C37F0D28E43D44EB5E27F8B23955E335D5DD3"))

	err = m.Read(bytes.NewBuffer(publishWrongWork))
	if blocks.ValidateBlockWork(m.ToBlock()) {
		t.Errorf("Invalid work should fail")
	}

	err = m.Read(bytes.NewBuffer(publishWrongSig))
	passed, _ := m.ToBlock().(*blocks.OpenBlock).VerifySignature()
	if passed {
		t.Errorf("Invalid signature should fail")
	}
}

func TestHandleMessage(t *testing.T) {
	store.Init(store.TestConfig)
	handleMessage(bytes.NewBuffer(publishTest))
}

func TestReadWriteHeader(t *testing.T) {
	var message MessageHeader
	buf := bytes.NewBuffer(publishOpen)
	message.ReadHeader(buf)

	if message.MagicNumber != MagicNumber {
		t.Errorf("Unexpected magic number")
	}

	if message.VersionMax != 4 {
		t.Errorf("Wrong VersionMax")
	}

	if message.VersionUsing != 5 {
		t.Errorf("Wrong VersionUsing")
	}

	if message.VersionMin != 1 {
		t.Errorf("Wrong VersionMin")
	}

	if message.MessageType != Message_publish {
		t.Errorf("Wrong Message Type")
	}

	if message.Extensions != 0 {
		t.Errorf("Wrong Extension")
	}

	if message.BlockType != BlockType_open {
		t.Errorf("Wrong Blocktype")
	}

	var writeBuf bytes.Buffer
	message.WriteHeader(&writeBuf)
	if bytes.Compare(publishOpen[:8], writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote header badly")
	}
}
