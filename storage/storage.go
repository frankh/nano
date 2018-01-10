package storage

import (
	"database/sql"
	"github.com/frankh/rai"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
	_ "github.com/mattn/go-sqlite3"
)

type Block interface {
	Type() blocks.BlockType
	GetBalance() uint128.Uint128
	Hash() rai.BlockHash
}

type OpenBlock struct {
	Source         SendBlock
	Representative rai.Account
	Account        rai.Account
	Work           rai.Work
	Signature      rai.Signature
}

type SendBlock struct {
	Previous    Block
	Balance     uint128.Uint128
	Destination rai.Account
	Work        rai.Work
	Signature   rai.Signature
}

type ReceiveBlock struct {
	Previous  Block
	Source    SendBlock
	Work      rai.Work
	Signature rai.Signature
}

type ChangeBlock struct {
	Previous       Block
	Representative rai.Account
	Work           rai.Work
	Signature      rai.Signature
}

func (b *OpenBlock) GetBalance() uint128.Uint128 {
	return b.Source.Previous.GetBalance().Sub(b.Source.GetBalance())
}

func (b *SendBlock) GetBalance() uint128.Uint128 {
	return b.Balance
}

func (b *ReceiveBlock) GetBalance() uint128.Uint128 {
	received := b.Source.Previous.GetBalance().Sub(b.Source.GetBalance())
	return b.Previous.GetBalance().Add(received)
}

func (b *ChangeBlock) GetBalance() uint128.Uint128 {
	return b.Previous.GetBalance()
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
        'account' TEXT NULL,
        'source' TEXT NULL,
        'representative' TEXT NULL,
        'previous' TEXT NULL,
        'balance' TEXT NULL,
        'destination' TEXT NULL,
        'work' TEXT NULL,
        'signature' VARCHAR(64) NULL,
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

	rows, err := Conn.Query(`SELECT hash FROM block WHERE hash=?`, blocks.LiveGenesisBlockHash)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		StoreBlock(blocks.LiveGenesisBlock)
	}

}

func StoreBlock(b blocks.RawBlock) {
	switch b.Type {
	case blocks.Open:
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

		_, err = prep.Exec(
			b.HashToString(),
			b.Source,
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
