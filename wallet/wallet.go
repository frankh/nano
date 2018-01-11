package wallet

import (
	"github.com/frankh/crypto/ed25519"
	"github.com/frankh/rai"
	"github.com/frankh/rai/address"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
	"github.com/pkg/errors"
)

type Wallet struct {
	privateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Head       blocks.Block
}

func New(private string) (w Wallet) {
	w.PublicKey, w.privateKey = address.KeypairFromPrivateKey(private)
	account := address.PubKeyToAddress(w.PublicKey)

	w.Head = blocks.FetchOpen(account)

	return w
}

func (w *Wallet) GetBalance() uint128.Uint128 {
	if w.Head == nil {
		return uint128.FromInts(0, 0)
	}

	return w.Head.GetBalance()

}

func (w *Wallet) Send(destination rai.Account, amount uint128.Uint128) (*blocks.RawBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot send from empty account")
	}

	block := blocks.NewSendBlock(
		w.Head.Hash(),
		w.GetBalance().Sub(amount),
		destination,
	)

	return &block, nil
}
