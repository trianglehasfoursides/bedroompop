package database

import (
	"path"

	"github.com/adrg/xdg"
	"go.etcd.io/bbolt"
)

func Get(name string, key []byte, value []byte) error {
	bbolt.Open(path.Join(xdg.DataHome, name+".bolt"), 0600, bbolt.DefaultOptions)
	return nil
}
