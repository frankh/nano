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
	var errs []error
	errs = append(errs,
		buf.WriteByte(m.MagicNumber[0]),
		buf.WriteByte(m.MagicNumber[1]),
		buf.WriteByte(m.VersionMax),
		buf.WriteByte(m.VersionUsing),
		buf.WriteByte(m.VersionMin),
		buf.WriteByte(m.MessageType),
		buf.WriteByte(m.Extensions),
		buf.WriteByte(m.BlockType),
	)

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MessageHeader) ReadHeader(buf *bytes.Buffer) error {
	var errs []error
	var err error
	// I really hate go error handling sometimes
	m.MagicNumber[0], err = buf.ReadByte()
	errs = append(errs, err)
	m.MagicNumber[1], err = buf.ReadByte()
	errs = append(errs, err)
	m.VersionMax, err = buf.ReadByte()
	errs = append(errs, err)
	m.VersionUsing, err = buf.ReadByte()
	errs = append(errs, err)
	m.VersionMin, err = buf.ReadByte()
	errs = append(errs, err)
	m.MessageType, err = buf.ReadByte()
	errs = append(errs, err)
	m.Extensions, err = buf.ReadByte()
	errs = append(errs, err)
	m.BlockType, err = buf.ReadByte()
	errs = append(errs, err)

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
