package sqlite

import (
	"sync"

	"github.com/trianglehasfoursides/mathrock/node/meta"
)

func deleteMetaDb(name string, mtx *sync.Mutex) {
	txn := meta.Meta.NewTransaction(true)
	mtx.Lock()
	defer mtx.Unlock()
	txn.Delete([]byte("meta:" + name))
	txn.Commit()
}
