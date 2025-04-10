package sqlite

import (
	_ "embed"
	"encoding/json"

	"github.com/dgraph-io/badger/v4"
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

func createConfigDb(name string) error {
	conf, _ := json.Marshal(config)
	configentry = badger.NewEntry([]byte("config:"+name), conf)
	txn = meta.Meta.NewTransaction(true)
	err = txn.SetEntry(configentry)
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}
