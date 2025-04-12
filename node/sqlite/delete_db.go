package sqlite

import (
	"errors"
	"os"
	"sync"

	"github.com/adrg/xdg"
)

// its delete the database
// i don't why am i writing thic comment
func DeleteDb(name string, mtx *sync.Mutex) error {
	if err := os.Remove(xdg.DataHome + "/" + name); err != nil {
		return errors.New("There is no database with the name: " + name)
	}
	deleteConfigDb(name, mtx)
	deleteMetaDb(name, mtx)
	return nil
}
