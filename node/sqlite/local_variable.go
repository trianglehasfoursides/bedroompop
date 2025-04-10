package sqlite

import (
	_ "embed"

	"github.com/dgraph-io/badger/v4"
)

type metaDb struct {
}

type configdb struct {
	size_limit   int64
	block_reads  bool
	block_writes bool
}

var (
	metadb      metaDb
	config      *configdb
	txn         *badger.Txn
	configentry *badger.Entry
	metaentry   *badger.Entry
	err         error
)
