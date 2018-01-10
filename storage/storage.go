package storage

import (
	"database/sql"
	"encoding/hex"
	"github.com/frankh/rai"
	"github.com/frankh/rai/blocks"
	"github.com/frankh/rai/uint128"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type Block interface {
	// Type() blocks.BlockType
	GetBalance() uint128.Uint128
	Hash() rai.BlockHash
}

type OpenBlock struct {
	SourceHash     rai.BlockHash
	Representative rai.Account
	Account        rai.Account
	Work           rai.Work
	Signature      rai.Signature
}

type SendBlock struct {
	PreviousHash rai.BlockHash
	Balance      uint128.Uint128
	Destination  rai.Account
	Work         rai.Work
	Signature    rai.Signature
}

type ReceiveBlock struct {
	PreviousHash rai.BlockHash
	SourceHash   rai.BlockHash
	Work         rai.Work
	Signature    rai.Signature
}

type ChangeBlock struct {
	PreviousHash   rai.BlockHash
	Representative rai.Account
	Work           rai.Work
	Signature      rai.Signature
}

func (b *OpenBlock) Hash() rai.BlockHash {
	return rai.BlockHash(strings.ToUpper(hex.EncodeToString(blocks.HashOpen(b.SourceHash, b.Representative, b.Account))))
}

func (b *ReceiveBlock) Hash() rai.BlockHash {
	return rai.BlockHash(strings.ToUpper(hex.EncodeToString(blocks.HashReceive(b.PreviousHash, b.SourceHash))))
}

func (b *ChangeBlock) Hash() rai.BlockHash {
	return rai.BlockHash(strings.ToUpper(hex.EncodeToString(blocks.HashChange(b.PreviousHash, b.Representative))))
}

func (b *SendBlock) Hash() rai.BlockHash {
	return rai.BlockHash(strings.ToUpper(hex.EncodeToString(blocks.HashSend(b.PreviousHash, b.Destination, b.Balance))))
}

func (b *OpenBlock) Source() *SendBlock {
	return FetchBlock(b.SourceHash).(*SendBlock)
}

func (b *ReceiveBlock) Source() *SendBlock {
	return FetchBlock(b.SourceHash).(*SendBlock)
}

func (b *ReceiveBlock) Previous() Block {
	return FetchBlock(b.PreviousHash)
}

func (b *ChangeBlock) Previous() Block {
	return FetchBlock(b.PreviousHash)
}

func (b *SendBlock) Previous() Block {
	return FetchBlock(b.PreviousHash)
}

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

	var block_type blocks.BlockType
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

	switch block_type {
	case blocks.Open:
		block := OpenBlock{
			source,
			representative,
			account,
			work,
			signature,
		}
		return &block
	default:
		panic("Unknown block type")
	}

}

func (b *OpenBlock) GetBalance() uint128.Uint128 {
	if b.SourceHash == blocks.LiveGenesisSourceHash {
		return blocks.LiveGenesisAmount
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
