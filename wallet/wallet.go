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

func (w *Wallet) Address() rai.Account {
	return address.PubKeyToAddress(w.PublicKey)
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

func (w *Wallet) Open(source rai.BlockHash, representative rai.Account, work *rai.Work) (*blocks.OpenBlock, error) {
	if w.Head != nil {
		return nil, errors.Errorf("Cannot open a non empty account")
	}

	existing := blocks.FetchOpen(w.Address())
	if existing != nil {
		return nil, errors.Errorf("Cannot open account, open block already exists")
	}

	send_block := blocks.FetchBlock(source)
	if send_block == nil {
		return nil, errors.Errorf("Could not find references send")
	}

	w.Work = work

	common := blocks.CommonBlock{
		*w.Work,
		"",
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

func (w *Wallet) Send(destination rai.Account, amount uint128.Uint128) (*blocks.SendBlock, error) {
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

	w.Head = &block
	return &block, nil
}

func (w *Wallet) Receive(source rai.BlockHash) (*blocks.ReceiveBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot receive to empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	send_block := blocks.FetchBlock(source)

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
		*w.Work,
		"",
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

func (w *Wallet) Change(representative rai.Account) (*blocks.ChangeBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot change on empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	common := blocks.CommonBlock{
		*w.Work,
		"",
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
