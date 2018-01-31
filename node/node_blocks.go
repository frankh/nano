package node

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/frankh/rai"
	"github.com/frankh/rai/address"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
	"github.com/frankh/rai/utils"
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

type MessageBlockChange struct {
	Previous       [32]byte
	Representative [32]byte
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

	work := make([]byte, 8)
	n, err = buf.Read(work)
	work = utils.Reversed(work)

	copy(m.Work[:], work)

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

	n, err = buf.Write(utils.Reversed(m.Work[:]))

	if n != len(m.Work) {
		return errors.New("Wrong number of bytes in work")
	}
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageBlockOpen) ToBlock() blocks.Block {
	common := blocks.CommonBlock{
		Work:      rai.Work(hex.EncodeToString(m.Work[:])),
		Signature: rai.Signature(hex.EncodeToString(m.Signature[:])),
	}

	block := blocks.OpenBlock{
		rai.BlockHash(hex.EncodeToString(m.Source[:])),
		address.PubKeyToAddress(m.Representative[:]),
		address.PubKeyToAddress(m.Account[:]),
		common,
	}

	return &block
}

func (m *MessageBlockSend) ToBlock() blocks.Block {
	common := blocks.CommonBlock{
		Work:      rai.Work(hex.EncodeToString(m.Work[:])),
		Signature: rai.Signature(hex.EncodeToString(m.Signature[:])),
	}

	block := blocks.SendBlock{
		rai.BlockHash(hex.EncodeToString(m.Previous[:])),
		address.PubKeyToAddress(m.Destination[:]),
		uint128.FromBytes(m.Balance[:]),
		common,
	}

	return &block
}

func (m *MessageBlockReceive) ToBlock() blocks.Block {
	common := blocks.CommonBlock{
		Work:      rai.Work(hex.EncodeToString(m.Work[:])),
		Signature: rai.Signature(hex.EncodeToString(m.Signature[:])),
	}

	block := blocks.ReceiveBlock{
		rai.BlockHash(hex.EncodeToString(m.Previous[:])),
		rai.BlockHash(hex.EncodeToString(m.Source[:])),
		common,
	}

	return &block
}

func (m *MessageBlockChange) ToBlock() blocks.Block {
	common := blocks.CommonBlock{
		Work:      rai.Work(hex.EncodeToString(m.Work[:])),
		Signature: rai.Signature(hex.EncodeToString(m.Signature[:])),
	}

	block := blocks.ChangeBlock{
		rai.BlockHash(hex.EncodeToString(m.Previous[:])),
		address.PubKeyToAddress(m.Representative[:]),
		common,
	}

	return &block
}

func (m *MessageBlockOpen) Read(buf *bytes.Buffer) error {
	n1, err1 := buf.Read(m.Source[:])
	n2, err2 := buf.Read(m.Representative[:])
	n3, err3 := buf.Read(m.Account[:])
	err4 := m.MessageBlockCommon.ReadCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return errors.New("Failed to read header")
	}

	if n1 != 32 || n2 != 32 || n3 != 32 {
		return errors.New("Wrong number of bytes read")
	}

	return nil
}

func (m *MessageBlockOpen) Write(buf *bytes.Buffer) error {
	n1, err1 := buf.Write(m.Source[:])
	n2, err2 := buf.Write(m.Representative[:])
	n3, err3 := buf.Write(m.Account[:])
	err4 := m.MessageBlockCommon.WriteCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return errors.New("Failed to write header")
	}

	if n1 != 32 || n2 != 32 || n3 != 32 {
		return errors.New("Wrong number of bytes written")
	}

	return nil
}

func (m *MessageBlockSend) Read(buf *bytes.Buffer) error {
	n1, err1 := buf.Read(m.Previous[:])
	n2, err2 := buf.Read(m.Destination[:])
	n3, err3 := buf.Read(m.Balance[:])
	err4 := m.MessageBlockCommon.ReadCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return errors.New("Failed to read header")
	}

	if n1 != 32 || n2 != 32 || n3 != 16 {
		return errors.New("Wrong number of bytes read")
	}

	return nil
}

func (m *MessageBlockSend) Write(buf *bytes.Buffer) error {
	n1, err1 := buf.Write(m.Previous[:])
	n2, err2 := buf.Write(m.Destination[:])
	n3, err3 := buf.Write(m.Balance[:])
	err4 := m.MessageBlockCommon.WriteCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return errors.New("Failed to write header")
	}

	if n1 != 32 || n2 != 32 || n3 != 16 {
		return errors.New("Wrong number of bytes written")
	}

	return nil
}

func (m *MessageBlockReceive) Read(buf *bytes.Buffer) error {
	n1, err1 := buf.Read(m.Previous[:])
	n2, err2 := buf.Read(m.Source[:])
	err3 := m.MessageBlockCommon.ReadCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("Failed to read header")
	}

	if n1 != 32 || n2 != 32 {
		return errors.New("Wrong number of bytes read")
	}

	return nil
}

func (m *MessageBlockReceive) Write(buf *bytes.Buffer) error {
	n1, err1 := buf.Write(m.Previous[:])
	n2, err2 := buf.Write(m.Source[:])
	err3 := m.MessageBlockCommon.WriteCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("Failed to write header")
	}

	if n1 != 32 || n2 != 32 {
		return errors.New("Wrong number of bytes written")
	}

	return nil
}

func (m *MessageBlockChange) Read(buf *bytes.Buffer) error {
	n1, err1 := buf.Read(m.Previous[:])
	n2, err2 := buf.Read(m.Representative[:])
	err3 := m.MessageBlockCommon.ReadCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("Failed to read change block")
	}

	if n1 != 32 || n2 != 32 {
		return errors.New("Wrong number of bytes read")
	}

	return nil
}

func (m *MessageBlockChange) Write(buf *bytes.Buffer) error {
	n1, err1 := buf.Write(m.Previous[:])
	n2, err2 := buf.Write(m.Representative[:])
	err3 := m.MessageBlockCommon.WriteCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("Failed to write change block")
	}

	if n1 != 32 || n2 != 32 {
		return errors.New("Wrong number of bytes written")
	}

	return nil
}
