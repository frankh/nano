package node

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var packet, _ = hex.DecodeString("5243040501030004FBC1F34CF9EF42FB137A909873BD3FDEC047CB8A6D4448B43C0610931E268F012298FAB7C61058E77EA554CB93EDEEDA0692CBFCC540AB213B2836B29029E23A0A3E8B35979AC58F7A0AB42656B28294F5968EB059749EA36BC372DDCDFDBB0134086DB608D63F4A086FD92E0BB4AC6A05926CEC84E4D7D99A86F81D90EA9669A9E02B4E907D5E09491206D76E4787F6F2C26B8FD9932315B10EC005A8B4F60DDA9D288B1C14A4CB")

func TestReadWriteMessage(t *testing.T) {
	var message MessagePublishOpen
	err := message.Read(bytes.NewBuffer(packet))

	if err != nil {
		t.Errorf("Failed to read message")
	}

	newPacket := make([]byte, len(packet))
	err = message.Write(bytes.NewBuffer(newPacket))

	if err != nil {
		t.Errorf("Failed to write message")
	}

	var writeBuf bytes.Buffer
	message.Write(&writeBuf)
	if bytes.Compare(packet, writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote message badly, expected %x saw %x", packet, newPacket)
	}
}

func TestReadWriteHeader(t *testing.T) {
	var message MessageHeader
	buf := bytes.NewBuffer(packet)
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
	if bytes.Compare(packet[:8], writeBuf.Bytes()) != 0 {
		t.Errorf("Wrote header badly")
	}
}
