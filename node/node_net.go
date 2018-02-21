package node

import (
	"bytes"
	"log"
	"math/rand"
	"net"
)

const packetSize = 512
const numberOfPeersToShare = 8

var DefaultPeer Peer
var PeerList []Peer
var PeerSet map[string]bool
var LocalIP string

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func ListenForUdp() {
	LocalIP = GetOutboundIP().String()

	log.Printf("Listening for udp packets on 7075")
	ln, err := net.ListenPacket("udp", ":7075")
	if err != nil {
		panic(err)
	}

	buf := make([]byte, packetSize)

	for {
		n, addr, err := ln.ReadFrom(buf)
		if err != nil {
			continue
		}

		source := addr.(*net.UDPAddr).IP.String()
		if n > 0 {
			handleMessage(source, bytes.NewBuffer(buf[:n]))
		}
	}
}

func SendKeepAlive(peer Peer) error {
	addr := peer.Addr()
	randomPeers := make([]Peer, 0, 2)

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

func SendKeepAlives(params []interface{}) {
	peers := PeerList
	for _, peer := range peers {
		// TODO: Handle errors
		SendKeepAlive(peer)
	}
}
