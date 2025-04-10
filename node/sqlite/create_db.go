package sqlite

import (
	"database/sql"
	"errors"
	"os"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

func CreateDb(name string) error {
	if err = os.Chdir(xdg.DataHome); err != nil {
		return err
	}
	if _, err = sql.Open("sqlite3", xdg.DataHome+"/"+name); err != nil {
		return err
	}
	if err = createMetaDb("meta:" + name); err != nil {
		err = errors.Join(os.Remove(xdg.DataHome+"/"+name), err)
		return err
	}
	if err = createConfigDb("config:" + name); err != nil {
		err = errors.Join(deleteMetaDb("meta:"+name), err)
		err = errors.Join(os.Remove(xdg.DataHome+"/"+name), err)
		return err
	}
	return nil
}
