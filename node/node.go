package node

import (
	"bytes"
	"errors"
)

var MagicNumber = [2]byte{'R', 'C'}

// Non-idiomatic constant names to keep consistent with reference implentation
const (
	Message_invalid uint8 = iota
	Message_not_a_type
	Message_keepalive
	Message_publish
	Message_confirm_req
	Message_confirm_ack
	Message_bulk_pull
	Message_bulk_push
	Message_frontier_req
)

const (
	BlockType_invalid uint8 = iota
	BlockType_not_a_block
	BlockType_send
	BlockType_receive
	BlockType_open
	BlockType_change
)

type MessageHeader struct {
	MagicNumber  [2]byte
	VersionMax   byte
	VersionUsing byte
	VersionMin   byte
	MessageType  byte
	Extensions   byte
	BlockType    byte
}

type MessageCommon struct {
	Signature [64]byte
	Work      [8]byte
}

type MessagePublishOpen struct {
	MessageHeader
	MessageBlockOpen
}

type MessagePublishSend struct {
	MessageHeader
	MessageBlockSend
}

func (m *MessageCommon) ReadCommon(buf *bytes.Buffer) error {
	n, err := buf.Read(m.Signature[:])

	if n != len(m.Signature) {
		return errors.New("Wrong number of bytes in signature")
	}
	if err != nil {
		return err
	}

	n, err = buf.Read(m.Work[:])

	if n != len(m.Work) {
		return errors.New("Wrong number of bytes in work")
	}
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageCommon) WriteCommon(buf *bytes.Buffer) error {
	n, err := buf.Write(m.Signature[:])

	if n != len(m.Signature) {
		return errors.New("Wrong number of bytes in signature")
	}
	if err != nil {
		return err
	}

	n, err = buf.Write(m.Work[:])

	if n != len(m.Work) {
		return errors.New("Wrong number of bytes in work")
	}
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageHeader) WriteHeader(buf *bytes.Buffer) error {
	buf.WriteByte(m.MagicNumber[0])
	buf.WriteByte(m.MagicNumber[1])
	buf.WriteByte(m.VersionMax)
	buf.WriteByte(m.VersionUsing)
	buf.WriteByte(m.VersionMin)
	buf.WriteByte(m.MessageType)
	buf.WriteByte(m.Extensions)
	buf.WriteByte(m.BlockType)
	return nil
}

func (m *MessageHeader) ReadHeader(buf *bytes.Buffer) error {
	m.MagicNumber[0], _ = buf.ReadByte()
	m.MagicNumber[1], _ = buf.ReadByte()
	m.VersionMax, _ = buf.ReadByte()
	m.VersionUsing, _ = buf.ReadByte()
	m.VersionMin, _ = buf.ReadByte()
	m.MessageType, _ = buf.ReadByte()
	m.Extensions, _ = buf.ReadByte()
	m.BlockType, _ = buf.ReadByte()
	return nil
}
