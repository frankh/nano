package store

import (
	"database/sql"
	"errors"
	"log"
	"sync"

	"github.com/frankh/rai"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Path         string
	GenesisBlock *blocks.OpenBlock
}

var LiveConfig = Config{
	"db.sqlite",
	blocks.LiveGenesisBlock,
}

var TestConfig = Config{
	":memory:",
	blocks.TestGenesisBlock,
}

var TestConfigLive = Config{
	":memory:",
	blocks.LiveGenesisBlock,
}

// Blocks that we cannot store due to not having their parent
// block stored
var unconnectedBlockPool map[rai.BlockHash]blocks.Block

var Conf *Config
var globalConn *sql.DB
var connLock sync.Mutex

func getConn() *sql.DB {
	connLock.Lock()
	if globalConn == nil {
		conn, err := sql.Open("sqlite3", Conf.Path)
		conn.SetMaxOpenConns(1)
		if err != nil {
			panic(err)
		}
		globalConn = conn
	}
	return globalConn
}

func releaseConn(conn *sql.DB) {
	connLock.Unlock()
}

func Init(config Config) {
	var err error
	unconnectedBlockPool = make(map[rai.BlockHash]blocks.Block)

	if globalConn != nil {
		globalConn.Close()
		globalConn = nil
	}
	Conf = &config
	conn := getConn()
	defer releaseConn(conn)

	rows, err := conn.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name='block'`)
	hasSchema := rows.Next()
	rows.Close()
	if err != nil {
		panic(err)
	}

	if !hasSchema {
		log.Println("Creating database schema")
		prep, err := conn.Prepare(`
      CREATE TABLE 'block' (
        'hash' TEXT PRIMARY KEY,
        'type' TEXT NOT NULL,
        'account' TEXT NOT NULL DEFAULT(''),
        'source' TEXT NOT NULL DEFAULT(''),
        'representative' TEXT NOT NULL DEFAULT(''),
        'previous' TEXT NOT NULL DEFAULT(''),
        'balance' TEXT NOT NULL DEFAULT(''),
        'destination' TEXT NOT NULL DEFAULT(''),
        'work' TEXT NOT NULL DEFAULT(''),
        'signature' TEXT NOT NULL DEFAULT(''),
        'created' DATE DEFAULT CURRENT_TIMESTAMP NOT NULL,
        'confirmed' INTEGER DEFAULT(0),
        FOREIGN KEY(previous) REFERENCES block(hash)
      );
      CREATE INDEX account_index ON block(account);
      CREATE INDEX type_index ON block(type);
    `)
		if err != nil {
			panic(err)
		}

		_, err = prep.Exec()
		if err != nil {
			panic(err)
		}
	}

	rows, err = conn.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name='block'`)
	hasSchema = rows.Next()
	rows.Close()
	if err != nil || !hasSchema {
		panic(err)
	}

	rows, err = conn.Query(`SELECT hash FROM block WHERE hash=?`, blocks.LiveGenesisBlockHash)
	hasGenesis := rows.Next()
	rows.Close()
	if err != nil {
		panic(err)
	}
	if !hasGenesis {
		uncheckedStoreBlock(conn, config.GenesisBlock)
	}
	log.Printf("Finished init db")

}

func FetchOpen(account rai.Account) (b *blocks.OpenBlock) {
	conn := getConn()
	defer releaseConn(conn)
	return fetchOpen(conn, account)
}

func fetchOpen(conn *sql.DB, account rai.Account) (b *blocks.OpenBlock) {
	rows, err := conn.Query(`SELECT
    type,
    source,
    representative,
    account,
    work,
    signature,
    previous,
    balance,
    destination
  FROM block WHERE type=? and account=?`, blocks.Open, account)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	if !rows.Next() {
		return nil
	}

	return blockFromRow(rows).(*blocks.OpenBlock)
}

func FetchBlock(hash rai.BlockHash) (b blocks.Block) {
	conn := getConn()
	defer releaseConn(conn)
	return fetchBlock(conn, hash)
}

func fetchBlock(conn *sql.DB, hash rai.BlockHash) (b blocks.Block) {
	rows, err := conn.Query(`SELECT
    type,
    source,
    representative,
    account,
    work,
    signature,
    previous,
    balance,
    destination
  FROM block WHERE hash=?`, hash)
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	if !rows.Next() {
		return nil
	}

	return blockFromRow(rows)
}

func blockFromRow(rows *sql.Rows) (b blocks.Block) {
	var block_type blocks.BlockType
	var source rai.BlockHash
	var representative rai.Account
	var account rai.Account
	var work rai.Work
	var signature rai.Signature
	var previous rai.BlockHash
	var balance string
	var destination rai.Account

	err := rows.Scan(
		&block_type,
		&source,
		&representative,
		&account,
		&work,
		&signature,
		&previous,
		&balance,
		&destination,
	)

	if err != nil {
		panic(err)
	}

	common := blocks.CommonBlock{
		Work:      work,
		Signature: signature,
	}

	switch block_type {
	case blocks.Open:
		block := blocks.OpenBlock{
			source,
			representative,
			account,
			common,
		}
		return &block
	case blocks.Send:
		balance_int, err := uint128.FromString(balance)
		if err != nil {
			panic(err)
		}
		block := blocks.SendBlock{
			previous,
			destination,
			balance_int,
			common,
		}
		return &block
	default:
		panic("Unknown block type")
	}
}

func GetBalance(block blocks.Block) uint128.Uint128 {
	conn := getConn()
	defer releaseConn(conn)
	return getBalance(conn, block)
}

func getSendAmount(conn *sql.DB, block *blocks.SendBlock) uint128.Uint128 {
	prev := fetchBlock(conn, block.PreviousHash)

	return getBalance(conn, prev).Sub(getBalance(conn, block))
}

func getBalance(conn *sql.DB, block blocks.Block) uint128.Uint128 {
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

func StoreBlock(block blocks.Block) error {
	conn := getConn()
	defer releaseConn(conn)
	return storeBlock(conn, block)
}

func storeBlock(conn *sql.DB, block blocks.Block) error {
	if !blocks.ValidateBlockWork(block) {
		return errors.New("Invalid work for block")
	}

	if block.Type() != blocks.Open && block.Type() != blocks.Change && block.Type() != blocks.Send && block.Type() != blocks.Receive {
		return errors.New("Unknown block type")
	}

	if fetchBlock(conn, block.RootHash()) == nil {
		unconnectedBlockPool[block.PreviousBlockHash()] = block
		log.Printf("Added block to unconnected pool, now %d", len(unconnectedBlockPool))
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

func uncheckedStoreBlock(conn *sql.DB, block blocks.Block) {
	switch block.Type() {
	case blocks.Open:
		prep, err := conn.Prepare(`
      INSERT INTO block (
        hash,
        type,
        source,
        representative,
        account,
        work,
        signature
      ) values (
        ?,
        'open',
        ?,
        ?,
        ?,
        ?,
        ?
      )
    `)

		if err != nil {
			panic(err)
		}

		b := block.(*blocks.OpenBlock)

		_, err = prep.Exec(
			b.Hash(),
			b.SourceHash,
			b.Representative,
			b.Account,
			b.Work,
			b.Signature,
		)

		if err != nil {
			panic(err)
		}
	case blocks.Send:
		prep, err := conn.Prepare(`
      INSERT INTO block (
        hash,
        type,
        previous,
        balance,
        destination,
        work,
        signature
      ) values (
        ?,
        'send',
        ?,
        ?,
        ?,
        ?,
        ?
      )
    `)

		if err != nil {
			panic(err)
		}

		b := block.(*blocks.SendBlock)

		_, err = prep.Exec(
			b.Hash(),
			b.PreviousHash,
			b.Balance.String(),
			b.Destination,
			b.Work,
			b.Signature,
		)

		if err != nil {
			panic(err)
		}
	case blocks.Receive:
		prep, err := conn.Prepare(`
      INSERT INTO block (
        hash,
        type,
        previous,
        source,
        work,
        signature
      ) values (
        ?,
        'receive',
        ?,
        ?,
        ?,
        ?
      )
    `)

		if err != nil {
			panic(err)
		}

		b := block.(*blocks.ReceiveBlock)

		_, err = prep.Exec(
			b.Hash(),
			b.PreviousHash,
			b.SourceHash,
			b.Work,
			b.Signature,
		)

		if err != nil {
			panic(err)
		}

	default:
		panic("Unknown block type")
	}
}
