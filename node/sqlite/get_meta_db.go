package sqlite

import (
	"errors"

	"github.com/dgraph-io/badger/v4"
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

func GetMetaDb(name string) (m []byte, err error) {
	txn = meta.Meta.NewTransaction(false)
	result, err := txn.Get([]byte("meta" + name))
	if err == badger.ErrKeyNotFound {
		return nil, errors.New("there is no database with the name: " + name)
	} else if err != nil {
		return nil, err
	}
	txn.Commit()
	result.Value(func(val []byte) error {
		m = val
		return nil
	})
	return m, nil
}
