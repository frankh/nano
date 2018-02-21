package cmd

import (
	"log"
	"math/rand"
	"net"
	"path"
	"time"

	"github.com/frankh/nano/node"
	"github.com/frankh/nano/store"

	"github.com/spf13/cobra"
)

var (
	InitialPeer string
	WorkDir     string
	TestNet     bool
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringVarP(&InitialPeer, "peer", "p", "::ffff:192.168.0.70", "Initial peer to make contact with")
	daemonCmd.Flags().StringVarP(&WorkDir, "work-dir", "d", "", "Directory to put generated files, e.g. db.")
	daemonCmd.Flags().BoolVarP(&TestNet, "testnet", "t", false, "Use test network configuration")
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts the node's daemon",
	Long:  `Starts a full Nano node as a long-running process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		node.DefaultPeer = node.Peer{
			net.ParseIP(InitialPeer),
			7075,
		}
		node.PeerList = []node.Peer{node.DefaultPeer}
		node.PeerSet = map[string]bool{node.DefaultPeer.String(): true}

		if TestNet {
			log.Println("Using test network configuration")
			store.TestConfig.Path = path.Join(WorkDir, store.TestConfig.Path)
			store.Init(store.TestConfig)
		} else {
			store.LiveConfig.Path = path.Join(WorkDir, store.LiveConfig.Path)
			store.Init(store.LiveConfig)
		}

		rand.Seed(time.Now().UnixNano())

		keepAliveSender := node.NewAlarm(node.AlarmFn(node.SendKeepAlives), []interface{}{}, 20*time.Second)
		node.ListenForUdp()

		keepAliveSender.Stop()

		return nil
	},
}
