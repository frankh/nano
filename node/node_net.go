package node

import (
	"bytes"
	"log"
	"math/rand"
	"net"
)

const packetSize = 512
const numberOfPeersToShare = 8

var PeerList = []Peer{Peer{
	net.ParseIP("::ffff:192.168.0.70"),
	7075,
}}

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
			log.Printf("Read message len %d: %x", n, buf[:n])
			handleMessage(bytes.NewBuffer(buf[:n]))
		}
	}
}

func SendKeepAlive(peer Peer) error {
	addr := peer.Addr()
	randomPeers := make([]Peer, 0)

	outConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	randIndices := rand.Perm(len(PeerList))
	for n, i := range randIndices {
		if n == numberOfPeersToShare {
			break
		}
		randomPeers = append(randomPeers, PeerList[i])
	}

	m := CreateKeepAlive(randomPeers)
	buf := bytes.NewBuffer(nil)
	m.Write(buf)

	outConn.Write(buf.Bytes())
	return nil
}
