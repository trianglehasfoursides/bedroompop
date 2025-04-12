package sqlite

import (
	"os"
	"strings"

	"github.com/adrg/xdg"
)

func ListDb() (names []string, err error) {
	entries, err := os.ReadDir(xdg.DataHome)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sqlite") {
			names = append(names, entry.Name())
		}
	}
	return
}

