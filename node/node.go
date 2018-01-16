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

type MessagePublishOpen struct {
	MessageHeader
	MessageBlockOpen
}

type MessagePublishSend struct {
	MessageHeader
	MessageBlockSend
}

type MessagePublishReceive struct {
	MessageHeader
	MessageBlockReceive
}

type MessagePublishChange struct {
	MessageHeader
	MessageBlockChange
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

func (m *MessagePublishOpen) Read(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.ReadHeader(buf)
	if m.MessageHeader.BlockType != BlockType_open {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockOpen.Read(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func (m *MessagePublishOpen) Write(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.WriteHeader(buf)
	if m.MessageHeader.BlockType != BlockType_open {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockOpen.Write(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func (m *MessagePublishSend) Read(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.ReadHeader(buf)
	if m.MessageHeader.BlockType != BlockType_send {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockSend.Read(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func (m *MessagePublishSend) Write(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.WriteHeader(buf)
	if m.MessageHeader.BlockType != BlockType_send {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockSend.Write(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}
func (m *MessagePublishReceive) Read(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.ReadHeader(buf)
	if m.MessageHeader.BlockType != BlockType_receive {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockReceive.Read(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func (m *MessagePublishReceive) Write(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.WriteHeader(buf)
	if m.MessageHeader.BlockType != BlockType_receive {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockReceive.Write(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func (m *MessagePublishChange) Read(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.ReadHeader(buf)
	if m.MessageHeader.BlockType != BlockType_change {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockChange.Read(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}

func (m *MessagePublishChange) Write(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.WriteHeader(buf)
	if m.MessageHeader.BlockType != BlockType_change {
		return errors.New("Wrong blocktype")
	}
	err2 := m.MessageBlockChange.Write(buf)

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	return nil
}
