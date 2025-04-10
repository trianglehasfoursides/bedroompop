package sqlite

import (
	"errors"

	"github.com/dgraph-io/badger/v4"
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

func GetConfigDb(name string) ([]byte, error) {
	var result []byte
	txn = meta.Meta.NewTransaction(false)
	item, err := txn.Get([]byte("config:" + name))
	if err == badger.ErrKeyNotFound {
		return nil, errors.New("oops i think you have the wrong key")
	} else if err != nil {
		return nil, err
	}
	item.Value(func(val []byte) error {
		result = val
		return nil
	})
	return result, nil
}
