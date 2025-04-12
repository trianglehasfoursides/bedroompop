package sqlite

import (
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

var metadata []byte

func getMetaDb(name string) []byte {
	txn := meta.Meta.NewTransaction(false)
	result, _ := txn.Get([]byte("meta:" + name))
	txn.Commit()
	result.Value(func(val []byte) error {
		metadata = val
		return nil
	})
	return metadata
}
