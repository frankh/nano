package storage

import (
	"database/sql"
	"github.com/frankh/rai/blocks"
	_ "github.com/mattn/go-sqlite3"
)

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

func StoreBlock(b blocks.Block) {
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
