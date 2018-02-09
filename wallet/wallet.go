package wallet

import (
	"encoding/hex"

	"github.com/frankh/crypto/ed25519"
	"github.com/frankh/nano/address"
	"github.com/frankh/nano/blocks"
	"github.com/frankh/nano/store"
	"github.com/frankh/nano/types"
	"github.com/frankh/nano/uint128"
	"github.com/pkg/errors"
)

type Wallet struct {
	privateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Head       blocks.Block
	Work       *types.Work
	PoWchan    chan types.Work
}

func (w *Wallet) Address() types.Account {
	return address.PubKeyToAddress(w.PublicKey)
}

func New(private string) (w Wallet) {
	w.PublicKey, w.privateKey = address.KeypairFromPrivateKey(private)
	account := address.PubKeyToAddress(w.PublicKey)

	open := store.FetchOpen(account)
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

func (w *Wallet) GeneratePowSync() error {
	err := w.GeneratePoWAsync()
	if err != nil {
		return err
	}

	w.WaitPoW()
	return nil
}

// Triggers a goroutine to generate the next proof of work.
func (w *Wallet) GeneratePoWAsync() error {
	if w.PoWchan != nil {
		return errors.Errorf("Already generating PoW")
	}

	w.PoWchan = make(chan types.Work)

	go func(c chan types.Work, w *Wallet) {
		if w.Head == nil {
			c <- blocks.GenerateWorkForHash(types.BlockHash(hex.EncodeToString(w.PublicKey)))
		} else {
			c <- blocks.GenerateWork(w.Head)
		}
	}(w.PoWchan, w)

	return nil
}

func (w *Wallet) GetBalance() uint128.Uint128 {
	if w.Head == nil {
		return uint128.FromInts(0, 0)
	}

	return store.GetBalance(w.Head)

}

func (w *Wallet) Open(source types.BlockHash, representative types.Account) (*blocks.OpenBlock, error) {
	if w.Head != nil {
		return nil, errors.Errorf("Cannot open a non empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	existing := store.FetchOpen(w.Address())
	if existing != nil {
		return nil, errors.Errorf("Cannot open account, open block already exists")
	}

	send_block := store.FetchBlock(source)
	if send_block == nil {
		return nil, errors.Errorf("Could not find references send")
	}

	common := blocks.CommonBlock{
		Work:      *w.Work,
		Signature: "",
	}

	block := blocks.OpenBlock{
		source,
		representative,
		w.Address(),
		common,
	}

	block.Signature = block.Hash().Sign(w.privateKey)

	if !blocks.ValidateBlockWork(&block) {
		return nil, errors.Errorf("Invalid PoW")
	}

	w.Head = &block
	return &block, nil
}

func (w *Wallet) Send(destination types.Account, amount uint128.Uint128) (*blocks.SendBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot send from empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	if amount.Compare(w.GetBalance()) > 0 {
		return nil, errors.Errorf("Tried to send more than balance")
	}

	common := blocks.CommonBlock{
		Work:      *w.Work,
		Signature: "",
	}

	block := blocks.SendBlock{
		w.Head.Hash(),
		destination,
		w.GetBalance().Sub(amount),
		common,
	}

	block.Signature = block.Hash().Sign(w.privateKey)

	w.Head = &block
	return &block, nil
}

func (w *Wallet) Receive(source types.BlockHash) (*blocks.ReceiveBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot receive to empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	send_block := store.FetchBlock(source)

	if send_block == nil {
		return nil, errors.Errorf("Source block not found")
	}

	if send_block.Type() != blocks.Send {
		return nil, errors.Errorf("Source block is not a send")
	}

	if send_block.(*blocks.SendBlock).Destination != w.Address() {
		return nil, errors.Errorf("Send is not for this account")
	}

	common := blocks.CommonBlock{
		Work:      *w.Work,
		Signature: "",
	}

	block := blocks.ReceiveBlock{
		w.Head.Hash(),
		source,
		common,
	}

	block.Signature = block.Hash().Sign(w.privateKey)

	w.Head = &block
	return &block, nil
}

func (w *Wallet) Change(representative types.Account) (*blocks.ChangeBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot change on empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	common := blocks.CommonBlock{
		Work:      *w.Work,
		Signature: "",
	}

	block := blocks.ChangeBlock{
		w.Head.Hash(),
		representative,
		common,
	}

	block.Signature = block.Hash().Sign(w.privateKey)

	w.Head = &block
	return &block, nil
}
