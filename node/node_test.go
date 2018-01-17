package node

import (
	"bytes"
	"encoding/hex"
	"github.com/frankh/rai"
	"github.com/frankh/rai/blocks"
	"testing"
)

var publishSend, _ = hex.DecodeString("5243050501030002B6460102018F076CC32FF2F65AD397299C47F8CA2BE784D5DE394D592C22BE8BFFBE91872F1D2A2BCC1CB47FB854D6D31E43C6391EADD5750BB9689E5DF0D6CB0000003D11C83DBCFF748EB4B7F7A3C059DDEEE5C8ECCC8F20DEF3AF3C4F0726F879082ED051D0C62A54CD69C4A66B020369B7033C5B0F77654173AB24D5C7A64CC4FFF0BDB368FCC989E41A656569047627C49A2A6D2FBC")
var publishReceive, _ = hex.DecodeString("5243050501030003233FF43F2ADE055D4D4BCC1C19A3100B720C21E5548A547B9B21938BBDBB19EE28A1763099135DADB3F223C0A4138269C7146A6431AF0597D24276BB0A24BAFCBA254A264BAA0BCBA5962A77E15D4EB021043FFFEA9E4391E179D467C66C69675E9634F9C124060FC65D5B2F67FCA38E8BA93BF910EB337010BC51E652B0640D62F2642CB37BCD7C")
var publishOpen, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC005A8B4F60DDA9D288B1C14A4CB")
var publishChange, _ = hex.DecodeString("5243050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED35BBAD0F1DE013836D3471C1")
var publishWrongMagic, _ = hex.DecodeString("5242050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED35BBAD0F1DE013836D3471C1")
var publishWrongSig, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC015A8B4F60DDA9D288B1C14A4CB")
var publishWrongWork, _ = hex.DecodeString("5242050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED34BBAD0F1DE013836D3471C0")

func TestReadWriteMessageBody(t *testing.T) {
	blocks.Init(blocks.TestConfig)
	var message MessagePublishOpen
	var header MessageHeader
	buf := bytes.NewBuffer(publishSend)
	header.ReadHeader(buf)
	message.MessageHeader = header

	// Try to read the wrong block type
	err := message.Read(buf)
	if err == nil {
		t.Errorf("Read send block as open")
	}

	buf = bytes.NewBuffer(publishOpen)
	header.ReadHeader(buf)
	message.MessageHeader = header
	err = message.Read(buf)
	if err != nil {
		t.Errorf("Failed to read message")
	}

	var writeBuf bytes.Buffer
	err = message.Write(&writeBuf)
	if err != nil {
		t.Errorf("Failed to write message")
	}

	if bytes.Compare(publishOpen[8:], writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly")
	}

	block := message.ToBlock().(*blocks.OpenBlock)
	if !blocks.ValidateBlockWork(block) {
		t.Errorf("Work validation failed")
	}

	if block.Account != "xrb_14jyjetsh8p7jxx1of38ctsa779okt9d1pdnmtjpqiukuq8zugr3bxpxf1zu" {
		t.Errorf("Deserialised account badly")
	}
}

func validateTestBlock(t *testing.T, b blocks.Block, expectedHash rai.BlockHash) {
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
	m, err := readMessagePublish(bytes.NewBuffer(publishSend))
	if err != nil {
		t.Errorf("Failed to read send message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), rai.BlockHash("687DCB9C8EB8AF9F39D8107C3432A8732EDBED1E3B5E2E0F6B86643D1EB5E24F"))

	m, err = readMessagePublish(bytes.NewBuffer(publishReceive))
	if err != nil {
		t.Errorf("Failed to read receive message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), rai.BlockHash("7D3E9D79342AA73B7148CB46706D23ED8BB0041A5316D67A053F336ABF0E6B60"))

	m, err = readMessagePublish(bytes.NewBuffer(publishOpen))
	if err != nil {
		t.Errorf("Failed to read open message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), rai.BlockHash("5F73CF212E58563734D57CCFCCEFE481DE40C96F097F594F4FA32C5585D84AA4"))

	m, err = readMessagePublish(bytes.NewBuffer(publishChange))
	if err != nil {
		t.Errorf("Failed to read change message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), rai.BlockHash("4AABA9923AC794B635B8C3CC275C37F0D28E43D44EB5E27F8B23955E335D5DD3"))

	m, err = readMessagePublish(bytes.NewBuffer(publishWrongWork))
	if blocks.ValidateBlockWork(m.ToBlock()) {
		t.Errorf("Invalid work should fail")
	}

	m, err = readMessagePublish(bytes.NewBuffer(publishWrongSig))
	passed, _ := m.ToBlock().(*blocks.OpenBlock).VerifySignature()
	if passed {
		t.Errorf("Invalid signature should fail")
	}
}

func TestReadWriteHeader(t *testing.T) {
	blocks.Init(blocks.TestConfig)
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
