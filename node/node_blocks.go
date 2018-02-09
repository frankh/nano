package node

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/frankh/nano/address"
	"github.com/frankh/nano/blocks"
	"github.com/frankh/nano/types"
	"github.com/frankh/nano/uint128"
	"github.com/frankh/nano/utils"
)

type MessageBlockCommon struct {
	Signature [64]byte
	Work      [8]byte
}

type MessageBlock struct {
	Type             byte
	SourceOrPrevious [32]byte // Source for open, previous for others
	RepDestOrSource  [32]byte // Rep for open/change, dest for send, source for receive
	Account          [32]byte // Account for open
	Balance          [16]byte // Balance for send
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

func (m *MessageBlock) ToBlock() blocks.Block {
	common := blocks.CommonBlock{
		Work:      types.Work(hex.EncodeToString(m.Work[:])),
		Signature: types.Signature(hex.EncodeToString(m.Signature[:])),
	}

	switch m.Type {
	case BlockType_open:
		block := blocks.OpenBlock{
			types.BlockHash(hex.EncodeToString(m.SourceOrPrevious[:])),
			address.PubKeyToAddress(m.RepDestOrSource[:]),
			address.PubKeyToAddress(m.Account[:]),
			common,
		}
		return &block
	case BlockType_send:
		block := blocks.SendBlock{
			types.BlockHash(hex.EncodeToString(m.SourceOrPrevious[:])),
			address.PubKeyToAddress(m.RepDestOrSource[:]),
			uint128.FromBytes(m.Balance[:]),
			common,
		}
		return &block
	case BlockType_receive:
		block := blocks.ReceiveBlock{
			types.BlockHash(hex.EncodeToString(m.SourceOrPrevious[:])),
			types.BlockHash(hex.EncodeToString(m.RepDestOrSource[:])),
			common,
		}
		return &block
	case BlockType_change:
		block := blocks.ChangeBlock{
			types.BlockHash(hex.EncodeToString(m.SourceOrPrevious[:])),
			address.PubKeyToAddress(m.RepDestOrSource[:]),
			common,
		}
		return &block
	default:
		return nil
	}
}

func (m *MessageBlock) Read(messageBlockType byte, buf *bytes.Buffer) error {
	m.Type = messageBlockType

	n1, err1 := buf.Read(m.SourceOrPrevious[:])
	n2, err2 := buf.Read(m.RepDestOrSource[:])

	if messageBlockType == BlockType_open {
		n, err := buf.Read(m.Account[:])
		if err != nil || n != 32 {
			return errors.New("Failed to read account")
		}
	}

	if messageBlockType == BlockType_send {
		n, err := buf.Read(m.Balance[:])
		if err != nil || n != 16 {
			return errors.New("Failed to read balance")
		}
	}

	err3 := m.MessageBlockCommon.ReadCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("Failed to read block")
	}

	if n1 != 32 || n2 != 32 {
		return errors.New("Wrong number of bytes read")
	}

	return nil
}

func (m *MessageBlock) Write(buf *bytes.Buffer) error {
	n1, err1 := buf.Write(m.SourceOrPrevious[:])
	n2, err2 := buf.Write(m.RepDestOrSource[:])

	if m.Type == BlockType_open {
		n, err := buf.Write(m.Account[:])
		if err != nil || n != 32 {
			return errors.New("Failed to write account")
		}
	}

	if m.Type == BlockType_send {
		n, err := buf.Write(m.Balance[:])
		if err != nil || n != 16 {
			return errors.New("Failed to write balance")
		}
	}

	err3 := m.MessageBlockCommon.WriteCommon(buf)

	if err1 != nil || err2 != nil || err3 != nil {
		return errors.New("Failed to write block")
	}

	if n1 != 32 || n2 != 32 {
		return errors.New("Wrong number of bytes written")
	}

	return nil
}
