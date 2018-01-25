package wallet

import (
	"encoding/hex"

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

func New(privKey string) (w Wallet) {
	w.PublicKey, w.privateKey = address.KeypairFromPrivateKey(privKey)
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

	w.PoWchan = make(chan rai.Work)

	go func(c chan rai.Work, w *Wallet) {
		if w.Head == nil {
			c <- blocks.GenerateWorkForHash(rai.BlockHash(hex.EncodeToString(w.PublicKey)))
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

	return w.Head.GetBalance()

}

func (w *Wallet) Open(source rai.BlockHash, representative rai.Account) (*blocks.OpenBlock, error) {
	if w.Head != nil {
		return nil, errors.Errorf("Cannot open a non empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	existing := blocks.FetchOpen(w.Address())
	if existing != nil {
		return nil, errors.Errorf("Cannot open account, open block already exists")
	}

	sendBlock := blocks.FetchBlock(source)
	if sendBlock == nil {
		return nil, errors.Errorf("Could not find references send")
	}

	commonBlock := blocks.CommonBlock{
		*w.Work,
		"",
	}

	openBlock := blocks.OpenBlock{
		source,
		representative,
		w.Address(),
		commonBlock,
	}

	openBlock.Signature = openBlock.Hash().Sign(w.privateKey)

	if !blocks.ValidateBlockWork(&openBlock) {
		return nil, errors.Errorf("Invalid PoW")
	}

	w.Head = &openBlock
	return &openBlock, nil
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

	commonBlock := blocks.CommonBlock{
		*w.Work,
		"",
	}

	sendBlock := blocks.SendBlock{
		w.Head.Hash(),
		destination,
		w.GetBalance().Sub(amount),
		commonBlock,
	}

	sendBlock.Signature = sendBlock.Hash().Sign(w.privateKey)

	w.Head = &sendBlock
	return &sendBlock, nil
}

func (w *Wallet) Receive(source rai.BlockHash) (*blocks.ReceiveBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot receive to empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	sendBlock := blocks.FetchBlock(source)

	if sendBlock == nil {
		return nil, errors.Errorf("Source block not found")
	}

	if sendBlock.Type() != blocks.Send {
		return nil, errors.Errorf("Source block is not a send")
	}

	if sendBlock.(*blocks.SendBlock).Destination != w.Address() {
		return nil, errors.Errorf("Send is not for this account")
	}

	commonBlock := blocks.CommonBlock{
		*w.Work,
		"",
	}

	receiveBlock := blocks.ReceiveBlock{
		w.Head.Hash(),
		source,
		commonBlock,
	}

	receiveBlock.Signature = receiveBlock.Hash().Sign(w.privateKey)

	w.Head = &receiveBlock
	return &receiveBlock, nil
}

func (w *Wallet) Change(representative rai.Account) (*blocks.ChangeBlock, error) {
	if w.Head == nil {
		return nil, errors.Errorf("Cannot change on empty account")
	}

	if w.Work == nil {
		return nil, errors.Errorf("No PoW")
	}

	commonBlock := blocks.CommonBlock{
		*w.Work,
		"",
	}

	changeBlock := blocks.ChangeBlock{
		w.Head.Hash(),
		representative,
		commonBlock,
	}

	changeBlock.Signature = changeBlock.Hash().Sign(w.privateKey)

	w.Head = &changeBlock
	return &changeBlock, nil
}
