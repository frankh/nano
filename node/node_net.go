package node

import (
	"bytes"
	"log"
	"math/rand"
	"net"
	"time"
)

const packetSize = 512
const numberOfPeersToShare = 8

var DefaultPeer = Peer{
	net.ParseIP("::ffff:192.168.0.70"),
	7075,
	nil,
}

var PeerList = []Peer{DefaultPeer}
var PeerSet = map[string]bool{DefaultPeer.String(): true}

func (p *Peer) SendMessage(m Message) error {
	now := time.Now()
	p.LastReachout = &now

	outConn, err := net.DialUDP("udp", nil, p.Addr())
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	err = m.Write(buf)
	if err != nil {
		return err
	}
	outConn.Write(buf.Bytes())

	return nil
}

func ListenForUdp() {
	log.Printf("Listening for udp packets on 7075")
	ln, err := net.ListenPacket("udp", ":7075")
	if err != nil {
		panic(err)
	}

	buf := make([]byte, packetSize)

	for {
		n, _, err := ln.ReadFrom(buf)
		if err != nil {
			continue
		}
		if n > 0 {
			handleMessage(bytes.NewBuffer(buf[:n]))
		}
	}
}

func SendKeepAlive(peer Peer) error {
	randomPeers := make([]Peer, 0)
	randIndices := rand.Perm(len(PeerList))
	for n, i := range randIndices {
		if n == numberOfPeersToShare {
			break
		}
		randomPeers = append(randomPeers, PeerList[i])
	}

	m := CreateKeepAlive(randomPeers)
	return peer.SendMessage(m)
}

func SendKeepAlives(params []interface{}) {
	peers := params[0].([]Peer)
	timeCutoff := time.Now().Add(-5 * time.Minute)

	for _, peer := range peers {
		if peer.LastReachout == nil || peer.LastReachout.Before(timeCutoff) {
			SendKeepAlive(peer)
		}
	}
}
