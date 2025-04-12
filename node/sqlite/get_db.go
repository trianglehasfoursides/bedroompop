package sqlite

import (
	"os"
)

func GetDb(name string) ([]byte, error) {
	if _, err := os.Stat(name); err != nil {
		return nil, err
	}
	db := append(getMetaDb(name), GetConfigDb(name)...)
	return db, nil
}
