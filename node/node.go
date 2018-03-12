package node

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/frankh/nano/store"
)

var MagicNumber = [2]byte{'R', 'C'}

const VersionMax = 0x05
const VersionUsing = 0x05
const VersionMin = 0x04

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

type Peer struct {
	IP           net.IP
	Port         uint16
	LastReachout *time.Time
}

type MessageHeader struct {
	MagicNumber  [2]byte
	VersionMax   byte
	VersionUsing byte
	VersionMin   byte
	MessageType  byte
	Extensions   byte
	BlockType    byte
}

type MessageKeepAlive struct {
	MessageHeader
	Peers []Peer
}

type MessageConfirmAck struct {
	MessageHeader
	MessageVote
}

type MessageConfirmReq struct {
	MessageHeader
	MessageBlock
}

type MessagePublish struct {
	MessageHeader
	MessageBlock
}

type Message interface {
	Write(buf *bytes.Buffer) error
}

func CreateKeepAlive(peers []Peer) *MessageKeepAlive {
	var m MessageKeepAlive
	m.MessageHeader.MagicNumber = MagicNumber
	m.MessageHeader.VersionMax = VersionMax
	m.MessageHeader.VersionUsing = VersionUsing
	m.MessageHeader.VersionMin = VersionMin
	m.MessageHeader.MessageType = Message_keepalive
	return &m
}

func (p *Peer) Addr() *net.UDPAddr {
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", p.IP.String(), p.Port))
	return addr
}

func (p *Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP.String(), p.Port)
}

func handleMessage(buf *bytes.Buffer) {
	var header MessageHeader
	header.ReadHeader(bytes.NewBuffer(buf.Bytes()))
	if header.MagicNumber != MagicNumber {
		log.Printf("Ignored message. Wrong magic number %s", header.MagicNumber)
		return
	}

	switch header.MessageType {
	case Message_keepalive:
		var m MessageKeepAlive
		err := m.Read(buf)
		if err != nil {
			log.Printf("Failed to read keepalive: %s", err)
		}
		log.Println("Read keepalive")
		err = m.Handle()
		if err != nil {
			log.Printf("Failed to handle keepalive")
		}
	case Message_publish:
		var m MessagePublish
		err := m.Read(buf)
		if err != nil {
			log.Printf("Failed to read publish: %s", err)
		} else {
			store.StoreBlock(m.ToBlock())
		}
	case Message_confirm_ack:
		var m MessageConfirmAck
		err := m.Read(buf)
		if err != nil {
			log.Printf("Failed to read confirm: %s", err)
		} else {
			store.StoreBlock(m.ToBlock())
		}
	default:
		log.Printf("Ignored message. Cannot handle message type %d\n", header.MessageType)
	}
}

func (m *MessageKeepAlive) Handle() error {
	for _, peer := range m.Peers {
		if !PeerSet[peer.String()] {
			PeerSet[peer.String()] = true
			PeerList = append(PeerList, peer)
			log.Printf("Added new peer to list: %s, now %d peers", peer.String(), len(PeerList))
		}
	}
	return nil
}

func (m *MessageKeepAlive) Read(buf *bytes.Buffer) error {
	var header MessageHeader
	err := header.ReadHeader(buf)
	if err != nil {
		return err
	}

	if header.MessageType != Message_keepalive {
		return errors.New("Tried to read wrong message type")
	}

	m.MessageHeader = header
	m.Peers = make([]Peer, 0)

	for {
		peerPort := make([]byte, 2)
		peerIp := make(net.IP, net.IPv6len)
		n, err := buf.Read(peerIp)
		if n == 0 {
			break
		}
		if err != nil {
			return err
		}
		n2, err := buf.Read(peerPort)
		if err != nil {
			return err
		}
		if n < net.IPv6len || n2 < 2 {
			return errors.New("Not enough ip bytes")
		}

		m.Peers = append(m.Peers, Peer{peerIp, binary.LittleEndian.Uint16(peerPort), nil})
	}

	return nil
}

func (m *MessageKeepAlive) Write(buf *bytes.Buffer) error {
	err := m.MessageHeader.WriteHeader(buf)
	if err != nil {
		return err
	}

	for _, peer := range m.Peers {
		_, err = buf.Write(peer.IP)
		if err != nil {
			return err
		}
		portBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(portBytes, peer.Port)
		if err != nil {
			return err
		}
		_, err = buf.Write(portBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MessageConfirmAck) Read(buf *bytes.Buffer) error {
	err := m.MessageHeader.ReadHeader(buf)
	if err != nil {
		return err
	}

	if m.MessageHeader.MessageType != Message_confirm_ack {
		return errors.New("Tried to read wrong message type")
	}
	err = m.MessageVote.Read(m.MessageHeader.BlockType, buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageConfirmAck) Write(buf *bytes.Buffer) error {
	err := m.MessageHeader.WriteHeader(buf)
	if err != nil {
		return err
	}

	err = m.MessageVote.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageConfirmReq) Read(buf *bytes.Buffer) error {
	err := m.MessageHeader.ReadHeader(buf)
	if err != nil {
		return err
	}

	if m.MessageHeader.MessageType != Message_confirm_req {
		return errors.New("Tried to read wrong message type")
	}
	err = m.MessageBlock.Read(m.MessageHeader.BlockType, buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageConfirmReq) Write(buf *bytes.Buffer) error {
	err := m.MessageHeader.WriteHeader(buf)
	if err != nil {
		return err
	}

	err = m.MessageBlock.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublish) Read(buf *bytes.Buffer) error {
	err := m.MessageHeader.ReadHeader(buf)
	if err != nil {
		return err
	}

	if m.MessageHeader.MessageType != Message_publish {
		return errors.New("Tried to read wrong message type")
	}
	err = m.MessageBlock.Read(m.MessageHeader.BlockType, buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *MessagePublish) Write(buf *bytes.Buffer) error {
	err := m.MessageHeader.WriteHeader(buf)
	if err != nil {
		return err
	}

	err = m.MessageBlock.Write(buf)
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
