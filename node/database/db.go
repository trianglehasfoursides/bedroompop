package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/bytedance/sonic"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// CreateDatabase creates a new database (SQLite, BoltDB, or DuckDB) along with its configuration.
func CreateDatabase(databaseName string, mutex *sync.Mutex) error {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+".sqlite")

	// Check if the database file already exists
	if _, err := os.Stat(databasePath); err == nil {
		return errors.New("the database already exists")
	}

	// Create SQLite database
	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return errors.Wrap(err, "failed to create SQLite database")
	}
	defer db.Close()

	// Save configuration for the database
	if err := SaveDatabaseConfiguration(databaseName, mutex); err != nil {
		_ = os.Remove(databasePath) // Clean up the database file if saving configuration fails
		return errors.Wrap(err, "failed to save database configuration")
	}

	return nil // Database created successfully
}

// DeleteDatabase deletes a database file (SQLite, BoltDB, or DuckDB) along with its configuration.
func DeleteDatabase(databaseName string, mutex *sync.Mutex) error {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+".sqlite")

	// Attempt to remove the database file
	if err := os.Remove(databasePath); err != nil {
		return errors.Wrap(err, "failed to delete database file")
	}

	// Remove associated configuration
	if err := DeleteDatabaseConfiguration(databaseName, mutex); err != nil {
		return errors.Wrap(err, "failed to delete database configuration")
	}

	return nil // Database deleted successfully
}

// GetDatabase retrieves the configuration for a specific database.
func GetDatabase(databaseName string) ([]byte, error) {
	// Construct the full path for the database file
	databasePath := filepath.Join(xdg.DataHome, databaseName+".sqlite")

	// Check if the database file exists
	if _, err := os.Stat(databasePath); err != nil {
		return nil, errors.New("database file does not exist")
	}

	// Retrieve the configuration
	configuration, err := GetDatabaseConfiguration(databaseName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve database configuration")
	}

	return configuration, nil // Return the configuration as a byte slice
}

// ListDatabases lists all SQLite database files in the data directory
func ListDatabase() ([]map[string]any, error) {
	// Read all entries in the data directory
	entries, err := os.ReadDir(xdg.DataHome)
	if err != nil {
		return nil, err // Return an error if reading the directory fails
	}

	// Filter entries to include only SQLite database files
	var databases []map[string]any
	for _, entry := range entries {
		// Extract the database name
		databaseName := strings.TrimSuffix(entry.Name(), ".sqlite")

		// Retrieve the configuration for the database
		configData, err := GetDatabaseConfiguration(databaseName)
		if err != nil {
			// If configuration retrieval fails, log the error and skip this database
			zap.L().Error("Failed to retrieve database configuration", zap.String("database", databaseName), zap.Error(err))
			continue
		}

		// Parse the configuration JSON into a struct
		var config DatabaseConfiguration
		if err := sonic.Unmarshal(configData, &config); err != nil {
			// If parsing fails, log the error and skip this database
			zap.L().Error("Failed to parse database configuration", zap.String("database", databaseName), zap.Error(err))
			continue
		}

		// Append the database and its configuration to the result
		databases = append(databases, map[string]any{
			"name":          databaseName,
			"configuration": config,
		})
	}

	return databases, nil // Return the list of databases with their configurations
}
