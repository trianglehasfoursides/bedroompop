package sqlite

import (
	"errors"
	"os"

	"github.com/adrg/xdg"
)

// its delete the database
// i don't why am i writing thic comment
func DeleteDb(name string) error {
	if err = os.Remove(xdg.DataHome + "/" + name); err != nil {
		return errors.New("There is no database with the name: " + name)
	}
	if err = errors.Join(deleteMetaDb("meta:"+name), deleteConfigDb("config:"+name), err); err != nil {
		return errors.New("the database is deleted,but you have a little problem:" + err.Error())
	}
	return nil
}
