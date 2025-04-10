package sqlite

import (
	_ "embed"
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

// the configuration fileee...
func createMetaDb(name string) error {
	metarawr, _ := json.Marshal(metadb)
	metaentry = badger.NewEntry([]byte("meta:"+name), metarawr)
	txn = meta.Meta.NewTransaction(true)
	err = txn.SetEntry(metaentry)
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}
