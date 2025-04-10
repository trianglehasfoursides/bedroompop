package sqlite

import "github.com/trianglehasfoursides/mathrock/node/meta"

func deleteConfigDb(name string) error {
	txn = meta.Meta.NewTransaction(true)
	if err = txn.Delete([]byte(name)); err != nil {
		return err
	}
	if err = txn.Commit(); err != nil {
		return err
	}
	return nil
}
