package node

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/frankh/rai"
	"github.com/frankh/rai/address"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
)

type MessageBlockCommon struct {
	Signature [64]byte
	Work      [8]byte
}

type MessageBlockOpen struct {
	Source         [32]byte
	Representative [32]byte
	Account        [32]byte
	MessageBlockCommon
}

type MessageBlockSend struct {
	Previous    [32]byte
	Destination [32]byte
	Balance     [16]byte
	MessageBlockCommon
}

type MessageBlockReceive struct {
	Previous [32]byte
	Source   [32]byte
	MessageBlockCommon
}

func (m *MessageBlockCommon) ReadCommon(buf *bytes.Buffer) error {
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

func (m *MessageBlockCommon) WriteCommon(buf *bytes.Buffer) error {
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

func (m *MessageBlockOpen) ToBlock() *blocks.OpenBlock {
	common := blocks.CommonBlock{
		rai.Work(hex.EncodeToString(m.Work[:])),
		rai.Signature(hex.EncodeToString(m.Signature[:])),
	}

	block := blocks.OpenBlock{
		rai.BlockHash(hex.EncodeToString(m.Source[:])),
		address.PubKeyToAddress(m.Representative[:]),
		address.PubKeyToAddress(m.Account[:]),
		common,
	}

	return &block
}

func (m *MessageBlockSend) ToBlock() *blocks.SendBlock {
	common := blocks.CommonBlock{
		rai.Work(hex.EncodeToString(m.Work[:])),
		rai.Signature(hex.EncodeToString(m.Signature[:])),
	}

	block := blocks.SendBlock{
		rai.BlockHash(hex.EncodeToString(m.Previous[:])),
		address.PubKeyToAddress(m.Destination[:]),
		uint128.FromBytes(m.Balance[:]),
		common,
	}

	return &block
}

func (m *MessagePublishOpen) Read(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.ReadHeader(buf)
	if m.MessageHeader.BlockType != BlockType_open {
		return errors.New("Wrong blocktype")
	}

	n2, err2 := buf.Read(m.Source[:])
	n3, err3 := buf.Read(m.Representative[:])
	n4, err4 := buf.Read(m.Account[:])
	err5 := m.MessageBlockCommon.ReadCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return errors.New("Failed to read header")
	}

	if n2 != 32 || n3 != 32 || n4 != 32 {
		return errors.New("Wrong number of bytes read")
	}

	return nil
}

func (m *MessagePublishOpen) Write(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.WriteHeader(buf)
	n2, err2 := buf.Write(m.Source[:])
	n3, err3 := buf.Write(m.Representative[:])
	n4, err4 := buf.Write(m.Account[:])
	err5 := m.MessageBlockCommon.WriteCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return errors.New("Failed to write header")
	}

	if n2 != 32 || n3 != 32 || n4 != 32 {
		return errors.New("Wrong number of bytes written")
	}

	return nil
}

func (m *MessagePublishSend) Read(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.ReadHeader(buf)
	if m.MessageHeader.BlockType != BlockType_send {
		return errors.New("Wrong blocktype")
	}

	n2, err2 := buf.Read(m.Previous[:])
	n3, err3 := buf.Read(m.Destination[:])
	n4, err4 := buf.Read(m.Balance[:])
	err5 := m.MessageBlockCommon.ReadCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return errors.New("Failed to read header")
	}

	if n2 != 32 || n3 != 32 || n4 != 16 {
		return errors.New("Wrong number of bytes read")
	}

	return nil
}

func (m *MessagePublishSend) Write(buf *bytes.Buffer) error {
	err1 := m.MessageHeader.WriteHeader(buf)
	n2, err2 := buf.Write(m.Previous[:])
	n3, err3 := buf.Write(m.Destination[:])
	n4, err4 := buf.Write(m.Balance[:])
	err5 := m.MessageBlockCommon.WriteCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return errors.New("Failed to write header")
	}

	if n2 != 32 || n3 != 32 || n4 != 16 {
		return errors.New("Wrong number of bytes written")
	}

	return nil
}
