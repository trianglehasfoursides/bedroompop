package sqlite

import (
	"encoding/json"

	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/mathrock/node/meta"
)

// TODO:pake for loop dan pake hashmap dan err untuk commit
func UpdateConfigDb(name string, vlu []byte) error {
	cfg := &configdb{}
	if err = json.Unmarshal(vlu, cfg); err != nil {
		return err
	}
	if gjson.Get(string(vlu), "size_limit").Exists() && gjson.Get(string(vlu), "size_limit").Int() < 50000 {
		cfg.size_limit = gjson.Get(string(vlu), "size_limit").Int()
	} else {
		cfg.size_limit = gjson.Get(string(getRawConfig(name)), "size_limit").Int()
	}
	if gjson.Get(string(vlu), "block_reads").Exists() {
		cfg.block_reads = gjson.Get(string(vlu), "block_reads").Bool()
	} else {
		cfg.block_reads = gjson.Get(string(getRawConfig(name)), "block_reads").Bool()
	}
	if gjson.Get(string(vlu), "block_writes").Exists() {
		cfg.block_writes = gjson.Get(string(vlu), "block_writes").Bool()
	} else {
		cfg.block_writes = gjson.Get(string(getRawConfig(name)), "block_reads").Bool()
	}

	vlu, err = json.Marshal(&cfg)
	if err != nil {
		return err
	}
	txn = meta.Meta.NewTransaction(true)
	if err = txn.Set([]byte("config:"+name), vlu); err != nil {
		return err
	}
	txn.Commit()
	return nil
}
