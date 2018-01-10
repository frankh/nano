package blocks

import (
	"database/sql"
	"github.com/frankh/rai"
	"github.com/frankh/rai/uint128"
	_ "github.com/mattn/go-sqlite3"
)

func FetchBlock(hash rai.BlockHash) (b Block) {
	if Conn == nil {
		panic("Database connection not initialised")
	}

	rows, err := Conn.Query(`SELECT
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

	if err != nil {
		panic(err)
	}

	if !rows.Next() {
		return nil
	}

	var block_type BlockType
	var source rai.BlockHash
	var representative rai.Account
	var account rai.Account
	var work rai.Work
	var signature rai.Signature
	var previous rai.BlockHash
	var balance string
	var destination rai.Account

	err = rows.Scan(
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

	common := CommonBlock{
		work,
		signature,
	}

	switch block_type {
	case Open:
		block := OpenBlock{
			source,
			representative,
			account,
			common,
		}
		return &block
	default:
		panic("Unknown block type")
	}

}

func (b *OpenBlock) GetBalance() uint128.Uint128 {
	if b.SourceHash == LiveGenesisSourceHash {
		return LiveGenesisAmount
	}

	return b.Source().Previous().GetBalance().Sub(b.Source().GetBalance())
}

func (b *SendBlock) GetBalance() uint128.Uint128 {
	return b.Balance
}

func (b *ReceiveBlock) GetBalance() uint128.Uint128 {
	received := b.Source().Previous().GetBalance().Sub(b.Source().GetBalance())
	return b.Previous().GetBalance().Add(received)
}

func (b *ChangeBlock) GetBalance() uint128.Uint128 {
	return b.Previous().GetBalance()
}

var Conn *sql.DB

func Init(path string) {
	var err error

	Conn, err = sql.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}

	table_check, err := Conn.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name='block'`)
	if err != nil {
		panic(err)
	}

	if !table_check.Next() {
		prep, err := Conn.Prepare(`
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
        FOREIGN KEY(previous) REFERENCES block(hash)
      )
    `)
		if err != nil {
			panic(err)
		}

		_, err = prep.Exec()
		if err != nil {
			panic(err)
		}
	}

	rows, err := Conn.Query(`SELECT hash FROM block WHERE hash=?`, LiveGenesisBlockHash)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		StoreBlock(LiveGenesisBlock)
	}

}

func StoreBlock(block Block) {
	switch block.Type() {
	case Open:
		prep, err := Conn.Prepare(`
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

		b := block.(*OpenBlock)

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
	default:
		panic("Unknown block type")
	}
}
