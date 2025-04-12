package sqlite

import (
	_ "embed"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

type configDb struct {
	SizeLimit   int64 `json:"size_limit,omitempty"`
	BlockReads  bool  `json:"block_reads,omitempty"`
	BlockWrites bool  `json:"block_writes,omitempty"`
}

var config *configDb

func createConfigDb(name string, mtx *sync.Mutex) {
	conf, _ := sonic.Marshal(config)
	txn := meta.Meta.NewTransaction(true)
	mtx.Lock()
	defer mtx.Unlock()
	txn.Set([]byte(name), conf)
	txn.Commit()
}
