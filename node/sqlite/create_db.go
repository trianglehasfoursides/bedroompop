package sqlite

import (
	"database/sql"
	"errors"
	"os"
	"sync"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

func CreateDb(name string, mtx *sync.Mutex) error {
	if _, err := os.Stat(name); err != nil {
		db, err := sql.Open("sqlite3", xdg.DataHome+"/"+name+".sqlite")
		if err != nil {
			return err
		}
		db.Close()
		createMetaDb(name, mtx)
		createConfigDb(name, mtx)
	}
	return errors.New("the database already exist")
}
