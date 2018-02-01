package node

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/frankh/nano"
	"github.com/frankh/nano/blocks"
	"github.com/frankh/nano/store"
)

var publishSend, _ = hex.DecodeString("5243050501030002B6460102018F076CC32FF2F65AD397299C47F8CA2BE784D5DE394D592C22BE8BFFBE91872F1D2A2BCC1CB47FB854D6D31E43C6391EADD5750BB9689E5DF0D6CB0000003D11C83DBCFF748EB4B7F7A3C059DDEEE5C8ECCC8F20DEF3AF3C4F0726F879082ED051D0C62A54CD69C4A66B020369B7033C5B0F77654173AB24D5C7A64CC4FFF0BDB368FCC989E41A656569047627C49A2A6D2FBC")
var publishReceive, _ = hex.DecodeString("5243050501030003233FF43F2ADE055D4D4BCC1C19A3100B720C21E5548A547B9B21938BBDBB19EE28A1763099135DADB3F223C0A4138269C7146A6431AF0597D24276BB0A24BAFCBA254A264BAA0BCBA5962A77E15D4EB021043FFFEA9E4391E179D467C66C69675E9634F9C124060FC65D5B2F67FCA38E8BA93BF910EB337010BC51E652B0640D62F2642CB37BCD7C")
var publishTest, _ = hex.DecodeString("52430505010300030AFC4456F1A54722B101E41B1C2E3F7AF0EFD456EAE3621786C021D72C0BA9880FD491C3FF52227C8CDF76C88CE8F650320042349210AD2681134FD74080675C60734FAA7F89DDF5BDA156A5C7996A79F2CBD22E244B4E39D497261D356A30BE70973313A71A7D52700A560191B8A926FCE44B987A96FE61A8C469BBE383340831783CA6A6511D6A")
var publishOpen, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC005A8B4F60DDA9D288B1C14A4CB")
var publishChange, _ = hex.DecodeString("5243050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED35BBAD0F1DE013836D3471C1")
var publishWrongMagic, _ = hex.DecodeString("5242050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED35BBAD0F1DE013836D3471C1")
var publishWrongSig, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC015A8B4F60DDA9D288B1C14A4CB")
var publishWrongWork, _ = hex.DecodeString("5242050501030005611A6FA8736497E6C1BD9AE42090F0F646F56B32B6E02F804C2295B3888A2FEDE196157A3B52034755CA905AD0C365B192A40203D8983E077093BCD6C9757A64A772CD1736F8DF3C6E382BDC7EED1D48628A65263CE50B12A603B6782D2C3E5EE2280B3C97ACEA67FF003CA3690B2BBEE160E375D0CAA220109D63ED34BBAD0F1DE013836D3471C0")
var keepAlive, _ = hex.DecodeString("524305050102000000000000000000000000FFFF49B13E26A31B00000000000000000000FFFF637887DF340400000000000000000000FFFFCC2C6D15A31B00000000000000000000FFFF5EC16857239C00000000000000000000FFFF23BD2D1FA31B00000000000000000000FFFF253B710AA31B00000000000000000000FFFF50740256A7E500000000000000000000FFFF4631D644A31B")

var confirmAck1, _ = hex.DecodeString("524305050105000289aaf8e5f19f60ebc9476f382dbee256deae2695b47934700d9aad49d86ccb249ceb5c2840fe3fdf2dcb9c40e142181e7bd158d07ca3f8388dc3b3c0acd395d85b38e04ce1dac45b070957046d31eb7f58caaa777a5e13d85fe2aae7514b490e9c1dd00100000000aef053ab1832d41df356290a704e6c6c47787c6da4710ee2399e60e0ab607e9e51380a2c22710ed4018392474228b4e7c80f1c6714dcc3c9ef4befa563ecc35905bd9a62bd5b7ebdc5ebc9f576392e00445a07742dc4b2bc1355aef245522b19ae5640985f7759954ebf5147a125fec7e9f1973cf1d2a9d182c9223392b4cc10cdb11bca27c455ec8b13f4482b506d02576cfad0046c5f1c")
var confirmAck2, _ = hex.DecodeString("52430505010500030c6d2dab2d2926b74b212d4f3ba3495c761b5fc9caceb017bea14826c7a29a6aa521fe6f5ddb7c8b1d44e849130cf388ca69fc01f1694ded1bd63825a09677bd57667e64a715dd2f24846e0a64615a381189c9613e0b092263c459d7767ef00c752ab80000000000e36cfed3204caa3684fe73cdb5b315c13616996b47d79891073b9d9ede2812a15230771c276fddf8917befebf6c3059735482bb49eb4c9fd90abd5509e24ffc555e8fbeeaac41a7069abaa14ee5f6c2e05c7024895be1f583471d860524aaba6257683f4f58ca75a6ab88897e58afdad5cd0ef88aa091996aa8e985e3fbc730e69a24966c33baf09")
var confirmAck3, _ = hex.DecodeString("52430505010500022b99e235183671a4f5b1eaacb4b8a410532df5bdca1ff47b973f71d597838884221fd91f7562af0d28b07141ada22de39fbc1230ede7c485185f4cf856f66f34ad8d2b0a7f073c388eb5728741822b0955e8ab65c4fcdbe5b5e0175410fe61058ca56d020000000048f071e85774719104ae7a1ed686697cab3f3dfd82b6cd769cfab25ff88bde3864a240ef387aec35fb7886ce0762855237a10663d7a1cbe27676fb233dfcbbfe0032d274116365eac8b9577c9440000089e448d4171372f0c11a1f175f90b3ae96aa9caa40f03081d1f159a5d6eff947db0dd618c151eb669443f974da483f49826131cb86a7690750fa9ece09ed1a03b2ad17b8e33a5e58")
var confirmAck4, _ = hex.DecodeString("5243050501050002922d41510187def9064712be5f2ef089b724bcb08f1e7f616623b6f0e68fcb40b7ec32fe6ca58ab5c4aa75db254f1490a9e88d0cbd1007ed6dc48bae26c5effc312f7c725587602c658db75c2821fe3c17c17d8645eedc390e12db9b05c960096fb5d5010000000048f071e85774719104ae7a1ed686697cab3f3dfd82b6cd769cfab25ff88bde3864a240ef387aec35fb7886ce0762855237a10663d7a1cbe27676fb233dfcbbfe0032d274116365eac8b9577c9440000089e448d4171372f0c11a1f175f90b3ae96aa9caa40f03081d1f159a5d6eff947db0dd618c151eb669443f974da483f49826131cb86a7690750fa9ece09ed1a03b2ad17b8e33a5e58")
var confirmAck5, _ = hex.DecodeString("52430505010500033f8ab2f1d088c2a35ca79583c249b9f0b20dd97b4d42b01c4862fcc33f45aeb4ee5d5141c8577a3e0b0365c80a1870c1afb472f2f341c2eec600b9487fcbd04a8bc4bc17327c07e5ef211c2def13ece34098518ede7261311a9263210f7d300c9887470000000000dd1c2240c918c66c19042fa9c68455546fb5e8dbab9790f338571c5e79af392031580d4a0fc9cfb96ca24e7f9d115c658513dd6ac6da450887720dbc91a26209d85d65c5e330f69247eeec77b212529b17c4861fc3593dd02617a1425a01a117cce72b3cf7aae5d9980953dfc70520bf33e3cb93e5f02197b2d4cbbb6ab4f00486828801cf0145e2")

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

func TestReadWriteMessageBody(t *testing.T) {
	store.Init(store.TestConfig)

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

	if bytes.Compare(publishOpen[8:], writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly")
	}

	block := m.ToBlock().(*blocks.OpenBlock)
	if !blocks.ValidateBlockWork(block) {
		t.Errorf("Work validation failed")
	}

	if block.Account != "xrb_14jyjetsh8p7jxx1of38ctsa779okt9d1pdnmtjpqiukuq8zugr3bxpxf1zu" {
		t.Errorf("Deserialised account badly")
	}
}

func validateTestBlock(t *testing.T, b blocks.Block, expectedHash nano.BlockHash) {
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
	validateTestBlock(t, m.ToBlock(), nano.BlockHash("687DCB9C8EB8AF9F39D8107C3432A8732EDBED1E3B5E2E0F6B86643D1EB5E24F"))

	err = m.Read(bytes.NewBuffer(publishReceive))
	if err != nil {
		t.Errorf("Failed to read receive message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), nano.BlockHash("7D3E9D79342AA73B7148CB46706D23ED8BB0041A5316D67A053F336ABF0E6B60"))

	err = m.Read(bytes.NewBuffer(publishOpen))
	if err != nil {
		t.Errorf("Failed to read open message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), nano.BlockHash("5F73CF212E58563734D57CCFCCEFE481DE40C96F097F594F4FA32C5585D84AA4"))

	err = m.Read(bytes.NewBuffer(publishChange))
	if err != nil {
		t.Errorf("Failed to read change message %s", err)
	}
	validateTestBlock(t, m.ToBlock(), nano.BlockHash("4AABA9923AC794B635B8C3CC275C37F0D28E43D44EB5E27F8B23955E335D5DD3"))

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
	store.Init(store.TestConfig)
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
