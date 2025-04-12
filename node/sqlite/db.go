package sqlite

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/bytedance/sonic"
	_ "github.com/mattn/go-sqlite3"
)

// InitializeDatabase creates a new SQLite database along with its configuration
func InitializeDatabase(databaseName string, mutex *sync.Mutex) error {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+".sqlite")

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
	defer db.Close() // Ensure the database connection is closed

	// Save configuration for the database
	if err := SaveDatabaseConfiguration(databaseName, mutex); err != nil {
		return err // Return an error if saving configuration fails
	}

	return nil // Return nil if the database is created successfully
}

// RemoveDatabase deletes the SQLite database file along with its configuration
func RemoveDatabase(databaseName string, mutex *sync.Mutex) error {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+".sqlite")

	// Attempt to remove the database file
	if err := os.Remove(databasePath); err != nil {
		return errors.New("no database found with the name: " + databaseName)
	}

	// Remove associated configuration
	if err := DeleteDatabaseConfiguration(databaseName, mutex); err != nil {
		return err // Return an error if deleting configuration fails
	}

	return nil // Return nil if the database is removed successfully
}

// RetrieveDatabase retrieves the configuration for a specific database
func RetrieveDatabase(databaseName string) ([]byte, error) {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+".sqlite")

	// Check if the database file exists
	if _, err := os.Stat(databasePath); err != nil {
		return nil, err // Return an error if the database file does not exist
	}

	// Retrieve the configuration
	configuration, err := GetDatabaseConfiguration(databaseName)
	if err != nil {
		return nil, err // Return an error if retrieving configuration fails
	}

	return configuration, nil // Return the configuration as a byte slice
}

// ListDatabases lists all SQLite database files in the data directory
func ListDatabases() ([]map[string]interface{}, error) {
	// Read all entries in the data directory
	entries, err := os.ReadDir(xdg.DataHome)
	if err != nil {
		return nil, err // Return an error if reading the directory fails
	}

	// Filter entries to include only SQLite database files
	var databases []map[string]interface{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sqlite") {
			// Extract the database name (without the .sqlite extension)
			databaseName := strings.TrimSuffix(entry.Name(), ".sqlite")

			// Retrieve the configuration for the database
			configData, err := GetDatabaseConfiguration(databaseName)
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
			databases = append(databases, map[string]interface{}{
				"name":          databaseName,
				"configuration": config,
			})
		}
	}

	return databases, nil // Return the list of databases with their configurations
}
