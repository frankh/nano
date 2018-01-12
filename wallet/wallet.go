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
	Work       *rai.Work
	PoWchan    chan rai.Work
}

func New(private string) (w Wallet) {
	w.PublicKey, w.privateKey = address.KeypairFromPrivateKey(private)
	account := address.PubKeyToAddress(w.PublicKey)

	open := blocks.FetchOpen(account)
	if open != nil {
		w.Head = open
	}

	return w
}

// Returns true if the wallet has prepared proof of work,
func (w *Wallet) HasPoW() bool {
	select {
	case work := <-w.PoWchan:
		w.Work = &work
		w.PoWchan = nil
		return true
	default:
		return false
	}
}

func (w *Wallet) WaitPoW() {
	for !w.HasPoW() {
	}
}

func (w *Wallet) WaitingForPoW() bool {
	return w.PoWchan != nil
}

// Triggers a goroutine to generate the next proof of work.
func (w *Wallet) GeneratePoWAsync() error {
	if w.PoWchan != nil {
		return errors.Errorf("Already generating PoW")
	}

	if w.Head == nil {
		return errors.Errorf("Cannot generate PoW on empty wallet")
	}

	w.PoWchan = make(chan rai.Work)

	go func(c chan rai.Work, b blocks.Block) {
		c <- blocks.GenerateWork(b)
	}(w.PoWchan, w.Head)

	return nil
}

func (w *Wallet) GetBalance() uint128.Uint128 {
	if w.Head == nil {
		return uint128.FromInts(0, 0)
	}

	return w.Head.GetBalance()

}

func (w *Wallet) Send(destination rai.Account, amount uint128.Uint128) (blocks.Block, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot send from empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	common := blocks.CommonBlock{
		*w.Work,
		"",
	}

	block := blocks.SendBlock{
		w.Head.Hash(),
		w.GetBalance().Sub(amount),
		destination,
		common,
	}

	block.Signature = block.Hash().Sign(w.privateKey)

	return &block, nil
}
