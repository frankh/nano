package node

import (
	"bytes"
	"encoding/hex"
	"github.com/frankh/rai"
	"github.com/frankh/rai/blocks"
	"testing"
)

var publishOpen, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC005A8B4F60DDA9D288B1C14A4CB")
var publishSend, _ = hex.DecodeString("5243050501030002B6460102018F076CC32FF2F65AD397299C47F8CA2BE784D5DE394D592C22BE8BFFBE91872F1D2A2BCC1CB47FB854D6D31E43C6391EADD5750BB9689E5DF0D6CB0000003D11C83DBCFF748EB4B7F7A3C059DDEEE5C8ECCC8F20DEF3AF3C4F0726F879082ED051D0C62A54CD69C4A66B020369B7033C5B0F77654173AB24D5C7A64CC4FFF0BDB368FCC989E41A656569047627C49A2A6D2FBC")

func TestReadWriteMessage(t *testing.T) {
	var message MessagePublishOpen

	// Try to read the wrong block type
	// As send blocks are smaller, pad with zeros
	// to ensure message header is being read
	err := message.Read(bytes.NewBuffer(append(publishSend, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)))
	if err == nil {
		t.Errorf("Read send block as open")
	}

	err = message.Read(bytes.NewBuffer(publishOpen))
	if err != nil {
		t.Errorf("Failed to read message")
	}

	var writeBuf bytes.Buffer
	err = message.Write(&writeBuf)
	if err != nil {
		t.Errorf("Failed to write message")
	}

	if bytes.Compare(publishOpen, writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly")
	}

	block := message.ToBlock()

	if !blocks.ValidateBlockWork(block) {
		t.Errorf("Work validation failed")
	}

	if block.Account != "xrb_14jyjetsh8p7jxx1of38ctsa779okt9d1pdnmtjpqiukuq8zugr3bxpxf1zu" {
		t.Errorf("Deserialised account badly")
	}
}

func TestMessagePublishSend(t *testing.T) {
	var message MessagePublishSend

	// Try to read the wrong block type
	err := message.Read(bytes.NewBuffer(publishOpen))
	if err == nil {
		t.Errorf("Read open block as send block")
	}

	err = message.Read(bytes.NewBuffer(publishSend))
	if err != nil {
		t.Errorf("Failed to read message")
	}

	var writeBuf bytes.Buffer
	err = message.Write(&writeBuf)
	if err != nil {
		t.Errorf("Failed to write message")
	}

	if bytes.Compare(publishSend, writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly")
	}

	block := message.ToBlock()

	if !blocks.ValidateBlockWork(block) {
		t.Errorf("Work validation failed")
	}

	if block.Hash() != rai.BlockHash("687DCB9C8EB8AF9F39D8107C3432A8732EDBED1E3B5E2E0F6B86643D1EB5E24F") {
		t.Errorf("Block hash is wrong")
	}

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
