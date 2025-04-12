package sqlite

import (
	_ "embed"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

type metaDb struct {
	Config *configDb
}

var metadb metaDb

func createMetaDb(name string, mtx *sync.Mutex) {
	metaa, _ := sonic.Marshal(metadb)
	txn := meta.Meta.NewTransaction(true)
	mtx.Lock()
	defer mtx.Unlock()
	txn.Set([]byte(name), metaa)
	txn.Commit()
}
