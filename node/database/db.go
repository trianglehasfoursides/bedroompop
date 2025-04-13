package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

// CreateDatabase creates a new SQLite database along with its configuration
func CreateDatabase(databaseName string, category string, mutex *sync.Mutex) error {
	switch category {
	case "sqlite":
		// Construct the full path for the database file
		databasePath := filepath.Join(xdg.DataHome, databaseName+"."+category)

		// Check if the database file already exists
		if _, err := os.Stat(databasePath); err == nil {
			// Return an error if the database already exists
			return errors.New("the database already exists")
		}

		// Attempt to create the database file
		db, err := sql.Open("sqlite3", databasePath)
		if err != nil {
			// Return an error if the database creation fails
			return err
		}

		// Save configuration for the database
		if err := SaveDatabaseConfiguration(databaseName, category, mutex); err != nil {
			return err // Return an error if saving configuration fails
		}

		db.Close() // Close teh database

		return nil // Return nil if the database is created successfully
	case "bolt":
		db, err := bbolt.Open(xdg.DataHome+databaseName+"."+category, 0600, nil)
		if err != nil {
			return err
		}
		defer db.Close()
		SaveDatabaseConfiguration(databaseName, category, mutex)
		return nil
	}
	return errors.New("no database was selected")
}
