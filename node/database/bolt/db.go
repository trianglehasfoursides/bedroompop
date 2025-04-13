package bolt

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"github.com/trianglehasfoursides/mathrock/node/database"
	"go.etcd.io/bbolt"
)

const (
	category = "bolt:"
	ext      = ".bolt"
)

func CreateDatabase(name string, mutex *sync.Mutex) (db *bbolt.DB, err error) {
	if db, err = bbolt.Open(xdg.DataHome+name+ext, 0600, nil); err != nil {
		return
	}
	defer db.Close()
	database.SaveDatabaseConfiguration(name, category, mutex)
	return
}

// RetrieveDatabase retrieves the configuration for a specific database
func GetDatabase(databaseName string) ([]byte, error) {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+ext)

	// Check if the database file exists
	if _, err := os.Stat(databasePath); err != nil {
		return nil, errors.New("") // Return an error if the database file does not exist
	}

	// Retrieve the configuration
	configuration, err := database.GetDatabaseConfiguration(databaseName, category)
	if err != nil {
		return nil, err // Return an error if retrieving configuration fails
	}

	return configuration, nil // Return the configuration as a byte slice
}

// RemoveDatabase deletes the SQLite database file along with its configuration
func RemoveDatabase(databaseName string, mutex *sync.Mutex) error {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+ext)

	// Attempt to remove the database file
	if err := os.Remove(databasePath); err != nil {
		return errors.New("no database found with the name: " + databaseName)
	}

	// Remove associated configuration
	if err := database.DeleteDatabaseConfiguration(databaseName, category, mutex); err != nil {
		return err // Return an error if deleting configuration fails
	}

	return nil // Return nil if the database is removed successfully
}
