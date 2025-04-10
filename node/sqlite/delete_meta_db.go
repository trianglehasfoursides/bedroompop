package sqlite

import (
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

func deleteMetaDb(name string) error {
	txn = meta.Meta.NewTransaction(true)
	err = txn.Delete([]byte("meta:" + name))
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}
