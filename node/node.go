package node

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/frankh/rai/blocks"
)

var MagicNumber = [2]byte{'R', 'C'}

// Non-idiomatic constant names to keep consistent with reference implentation
const (
	Message_invalid byte = iota
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
	BlockType_invalid byte = iota
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

type MessagePublish interface {
	Read(*bytes.Buffer) error
	Write(*bytes.Buffer) error
	ToBlock() blocks.Block
}

func messagePublishForHeader(header MessageHeader) (MessagePublish, error) {
	var m MessagePublish
	switch header.BlockType {
	case BlockType_send:
		var message MessagePublishSend
		message.MessageHeader = header
		m = &message
	case BlockType_receive:
		var message MessagePublishReceive
		message.MessageHeader = header
		m = &message
	case BlockType_open:
		var message MessagePublishOpen
		message.MessageHeader = header
		m = &message
	case BlockType_change:
		var message MessagePublishChange
		message.MessageHeader = header
		m = &message
	default:
		return nil, errors.New("Unknown block type")
	}

	return m, nil
}

func readMessagePublish(buf *bytes.Buffer) (MessagePublish, error) {
	var header MessageHeader
	err := header.ReadHeader(buf)
	if err != nil {
		return nil, err
	}

	if header.MessageType != Message_publish {
		return nil, errors.New("Tried to read wrong message type")
	}

	m, err := messagePublishForHeader(header)
	if err != nil {
		return nil, err
	}

	err = m.Read(buf)
	if err != nil {
		return nil, err
	}

	return m, nil
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
	if m.MessageHeader.BlockType != BlockType_open {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockOpen.Read(buf)

	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublishOpen) Write(buf *bytes.Buffer) error {
	if m.MessageHeader.BlockType != BlockType_open {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockOpen.Write(buf)

	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublishSend) Read(buf *bytes.Buffer) error {
	if m.MessageHeader.BlockType != BlockType_send {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockSend.Read(buf)

	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublishSend) Write(buf *bytes.Buffer) error {
	if m.MessageHeader.BlockType != BlockType_send {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockSend.Write(buf)

	if err != nil {
		return err
	}

	return nil
}
func (m *MessagePublishReceive) Read(buf *bytes.Buffer) error {
	if m.MessageHeader.BlockType != BlockType_receive {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockReceive.Read(buf)

	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublishReceive) Write(buf *bytes.Buffer) error {
	if m.MessageHeader.BlockType != BlockType_receive {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockReceive.Write(buf)

	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublishChange) Read(buf *bytes.Buffer) error {
	if m.MessageHeader.BlockType != BlockType_change {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockChange.Read(buf)

	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublishChange) Write(buf *bytes.Buffer) error {
	if m.MessageHeader.BlockType != BlockType_change {
		return errors.New(fmt.Sprintf("Wrong blocktype %d", m.MessageHeader.BlockType))
	}
	err := m.MessageBlockChange.Write(buf)

	if err != nil {
		return err
	}

	return nil
}
