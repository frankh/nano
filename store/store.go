package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/frankh/nano/address"
	"github.com/frankh/nano/blocks"
	"github.com/frankh/nano/types"
	"github.com/frankh/nano/uint128"
)

type Config struct {
	Path         string
	GenesisBlock *blocks.OpenBlock
}

const (
	MetaOpen byte = iota
	MetaReceive
	MetaSend
	MetaChange
)

type BlockItem struct {
	badger.Item
}

func (i *BlockItem) ToBlock() blocks.Block {
	meta := i.UserMeta()
	value, _ := i.Value()

	dec := gob.NewDecoder(bytes.NewBuffer(value))
	var result blocks.Block

	switch meta {
	case MetaOpen:
		var b blocks.OpenBlock
		dec.Decode(&b)
		result = &b
	case MetaReceive:
		var b blocks.ReceiveBlock
		dec.Decode(&b)
		result = &b
	case MetaSend:
		var b blocks.SendBlock
		dec.Decode(&b)
		result = &b
	case MetaChange:
		var b blocks.ChangeBlock
		dec.Decode(&b)
		result = &b
	}

	return result
}

var LiveConfig = Config{
	"DATA",
	blocks.LiveGenesisBlock,
}

var TestConfig = Config{
	"TESTDATA",
	blocks.TestGenesisBlock,
}

var TestConfigLive = Config{
	"TESTDATA",
	blocks.LiveGenesisBlock,
}

// Blocks that we cannot store due to not having their parent
// block stored
var unconnectedBlockPool map[types.BlockHash]blocks.Block

var Conf *Config
var globalConn *badger.DB
var currentTxn *badger.Txn
var connLock sync.Mutex

func getConn() *badger.Txn {
	connLock.Lock()

	if currentTxn != nil {
		return currentTxn
	}

	if globalConn == nil {
		opts := badger.DefaultOptions
		opts.Dir = Conf.Path
		opts.ValueDir = Conf.Path
		conn, err := badger.Open(opts)
		if err != nil {
			panic(err)
		}
		globalConn = conn
	}

	currentTxn = globalConn.NewTransaction(true)
	return currentTxn
}

func releaseConn(conn *badger.Txn) {
	currentTxn.Commit(nil)
	currentTxn = nil
	connLock.Unlock()
}

func Init(config Config) {
	var err error
	unconnectedBlockPool = make(map[types.BlockHash]blocks.Block)

	if globalConn != nil {
		globalConn.Close()
		globalConn = nil
	}
	Conf = &config
	conn := getConn()
	defer releaseConn(conn)

	_, err = conn.Get(blocks.LiveGenesisBlockHash.ToBytes())

	if err != nil {
		uncheckedStoreBlock(conn, config.GenesisBlock)
	}
}

func FetchOpen(account types.Account) (b *blocks.OpenBlock) {
	conn := getConn()
	defer releaseConn(conn)
	return fetchOpen(conn, account)
}

func fetchOpen(conn *badger.Txn, account types.Account) (b *blocks.OpenBlock) {
	account_bytes, err := address.AddressToPub(account)
	if err != nil {
		return nil
	}

	item, err := conn.Get(account_bytes)
	if err != nil {
		return nil
	}

	blockItem := BlockItem{*item}
	return blockItem.ToBlock().(*blocks.OpenBlock)
}

func FetchBlock(hash types.BlockHash) (b blocks.Block) {
	conn := getConn()
	defer releaseConn(conn)
	return fetchBlock(conn, hash)
}

func fetchBlock(conn *badger.Txn, hash types.BlockHash) (b blocks.Block) {
	item, err := conn.Get(hash.ToBytes())
	if err != nil {
		return nil
	}

	blockItem := BlockItem{*item}
	return blockItem.ToBlock()
}

func GetBalance(block blocks.Block) uint128.Uint128 {
	conn := getConn()
	defer releaseConn(conn)
	return getBalance(conn, block)
}

func getSendAmount(conn *badger.Txn, block *blocks.SendBlock) uint128.Uint128 {
	prev := fetchBlock(conn, block.PreviousHash)

	return getBalance(conn, prev).Sub(getBalance(conn, block))
}

func getBalance(conn *badger.Txn, block blocks.Block) uint128.Uint128 {
	switch block.Type() {
	case blocks.Open:
		b := block.(*blocks.OpenBlock)
		if b.SourceHash == Conf.GenesisBlock.SourceHash {
			return blocks.GenesisAmount
		}
		source := fetchBlock(conn, b.SourceHash).(*blocks.SendBlock)
		return getSendAmount(conn, source)

	case blocks.Send:
		b := block.(*blocks.SendBlock)
		return b.Balance

	case blocks.Receive:
		b := block.(*blocks.ReceiveBlock)
		prev := fetchBlock(conn, b.PreviousHash)
		source := fetchBlock(conn, b.SourceHash).(*blocks.SendBlock)
		received := getSendAmount(conn, source)
		return getBalance(conn, prev).Add(received)

	case blocks.Change:
		b := block.(*blocks.ChangeBlock)
		return getBalance(conn, fetchBlock(conn, b.PreviousHash))

	default:
		panic("Unknown block type")
	}

}

// Validate and store a block
// TODO: Validate signature and balance
func StoreBlock(block blocks.Block) error {
	conn := getConn()
	defer releaseConn(conn)
	return storeBlock(conn, block)
}

func storeBlock(conn *badger.Txn, block blocks.Block) error {
	if !blocks.ValidateBlockWork(block) {
		return errors.New("Invalid work for block")
	}

	if block.Type() != blocks.Open && block.Type() != blocks.Change && block.Type() != blocks.Send && block.Type() != blocks.Receive {
		return errors.New("Unknown block type")
	}

	if fetchBlock(conn, block.PreviousBlockHash()) == nil {
		if unconnectedBlockPool[block.PreviousBlockHash()] == nil {
			unconnectedBlockPool[block.PreviousBlockHash()] = block
			log.Printf("Added block to unconnected pool, now %d", len(unconnectedBlockPool))
		}
		return errors.New("Cannot find parent block")
	}

	uncheckedStoreBlock(conn, block)
	dependentBlock := unconnectedBlockPool[block.Hash()]

	if dependentBlock != nil {
		// We have an unconnected block dependent on this: Store it now that
		// it's connected
		delete(unconnectedBlockPool, block.Hash())
		storeBlock(conn, dependentBlock)
	}

	return nil
}

// Store a block without checking whether it's valid
// The block should be pre-checked to ensure it has a valid signature,
// parent block, balance, etc.
func uncheckedStoreBlock(conn *badger.Txn, block blocks.Block) {
	var buf bytes.Buffer
	var meta byte
	enc := gob.NewEncoder(&buf)
	switch block.Type() {
	case blocks.Open:
		b := block.(*blocks.OpenBlock)
		meta = MetaOpen
		err := enc.Encode(b)
		if err != nil {
			panic(err)
		}
		// Open blocks need to be stored twice, once keyed on account,
		// once keyed on hash.
		err = conn.SetWithMeta(b.RootHash().ToBytes(), buf.Bytes(), meta)
		if err != nil {
			panic(err)
		}
	case blocks.Send:
		b := block.(*blocks.SendBlock)
		meta = MetaSend
		err := enc.Encode(b)
		if err != nil {
			panic(err)
		}
	case blocks.Receive:
		b := block.(*blocks.ReceiveBlock)
		meta = MetaReceive
		err := enc.Encode(b)
		if err != nil {
			panic(err)
		}
	case blocks.Change:
		b := block.(*blocks.ChangeBlock)
		meta = MetaChange
		err := enc.Encode(b)
		if err != nil {
			panic(err)
		}
	default:
		panic("Unknown block type")
	}

	err := conn.SetWithMeta(block.Hash().ToBytes(), buf.Bytes(), meta)
	if err != nil {
		panic("Failed to store block")
	}
}
