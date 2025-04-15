package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/bytedance/sonic"
	_ "github.com/marcboeker/go-duckdb/v2"
	_ "github.com/mattn/go-sqlite3"
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
	case "duckdb":
		db, err := sql.Open("duckdb", "")
		if err != nil {
			return err
		}
		defer db.Close()
	case "pglite":
	case "sqlitex":
	case "postrock":
	}
	return errors.New("no database was selected")
}

// RemoveDatabase deletes the SQLite database file along with its configuration
func DeleteDatabase(databaseName string, category string, mutex *sync.Mutex) error {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+"."+category)

	// Attempt to remove the database file
	if err := os.Remove(databasePath); err != nil {
		return errors.New("no database found with the name: " + databaseName)
	}

	// Remove associated configuration
	if err := DeleteDatabaseConfiguration(databaseName, category, mutex); err != nil {
		return errors.New("error occurs while deleting database configuration : " + err.Error()) // Return an error if deleting configuration fails
	}

	return nil // Return nil if the database is removed successfully
}

// RetrieveDatabase retrieves the configuration for a specific database
func GetDatabase(databaseName string, category string) ([]byte, error) {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+"."+category)

	// Check if the database file exists
	if _, err := os.Stat(databasePath); err != nil {
		return nil, errors.New("") // Return an error if the database file does not exist
	}

	// Retrieve the configuration
	configuration, err := GetDatabaseConfiguration(databaseName, category)
	if err != nil {
		return nil, err // Return an error if retrieving configuration fails
	}

	return configuration, nil // Return the configuration as a byte slice
}

// ListDatabases lists all SQLite database files in the data directory
func ListDatabases(category string) ([]map[string]any, error) {
	// Read all entries in the data directory
	entries, err := os.ReadDir(xdg.DataHome)
	if err != nil {
		return nil, err // Return an error if reading the directory fails
	}

	// Filter entries to include only SQLite database files
	var databases []map[string]any
	for _, entry := range entries {
		if category != "" && strings.HasSuffix(entry.Name(), category) {
			// Extract the database name (without the .sqlite extension)
			databaseName := strings.Split(entry.Name(), ".")

			// Retrieve the configuration for the database
			configData, err := GetDatabaseConfiguration(databaseName[0], category)
			if err != nil {
				// If configuration retrieval fails, log the error and skip this database
				continue
			}

			// Parse the configuration JSON into a struct
			var config DatabaseConfiguration
			if err := sonic.Unmarshal(configData, &config); err != nil {
				// If parsing fails, log the error and skip this database
				continue
			}

			// Append the database and its configuration to the result
			databases = append(databases, map[string]any{
				"name":          databaseName,
				"configuration": config,
			})
		} else {
			// Extract the database name (without the .sqlite extension)
			databaseName := strings.TrimSuffix(entry.Name(), "."+category)

			// Retrieve the configuration for the database
			configData, err := GetDatabaseConfiguration(databaseName, category)
			if err != nil {
				// If configuration retrieval fails, log the error and skip this database
				continue
			}

			// Parse the configuration JSON into a struct
			var config DatabaseConfiguration
			if err := sonic.Unmarshal(configData, &config); err != nil {
				// If parsing fails, log the error and skip this database
				continue
			}

			// Append the database and its configuration to the result
			databases = append(databases, map[string]any{
				"name":          databaseName,
				"configuration": config,
			})
		}
	}

	return databases, nil // Return the list of databases with their configurations
}
