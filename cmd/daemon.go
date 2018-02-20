package cmd

import (
	"net"
	"time"

	"github.com/frankh/nano/node"
	"github.com/frankh/nano/store"

	"github.com/spf13/cobra"
)

var (
	InitialPeer string
	TestNet     bool
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringVarP(&InitialPeer, "peer", "p", "::ffff:192.168.0.70", "Initial peer to make contact with")
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

		if TestNet {
			store.Init(store.TestConfig)
		} else {
			store.Init(store.LiveConfig)
		}

		keepAliveSender := node.NewAlarm(node.AlarmFn(node.SendKeepAlives), []interface{}{node.PeerList}, 20*time.Second)
		node.ListenForUdp()

		keepAliveSender.Stop()

		return nil
	},
}
