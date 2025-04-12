package sqlite

import (
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

func GetConfigDb(name string) []byte {
	var result []byte
	txn := meta.Meta.NewTransaction(false)
	item, _ := txn.Get([]byte("config:" + name))
	item.Value(func(val []byte) error {
		result = val
		return nil
	})
	return result
}
