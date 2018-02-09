package main

import (
	"github.com/frankh/nano/node"
	"github.com/frankh/nano/store"
)

func main() {
	store.Init(store.LiveConfig)

	node.SendKeepAlive(node.PeerList[0])
	node.ListenForUdp()
}
