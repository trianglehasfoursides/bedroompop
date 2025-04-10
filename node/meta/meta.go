package meta

import (
	"log"

	"github.com/dgraph-io/badger/v4"
)

var (
	Meta = &badger.DB{}
	err  error
)

func init() {
	Meta, err = badger.Open(badger.DefaultOptions(""))
	if err != nil {
		log.Fatal("owww sorry mate,but the configuration db is not working,the error message is:" + err.Error())
	}
}
