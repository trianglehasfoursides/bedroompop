package database

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/dgraph-io/badger/v4"
)

// ConfigDB adalah instance global untuk database konfigurasi
var (
	ConfigDB *badger.DB // Instance dari BadgerDB untuk menyimpan metadata konfigurasi
	dbErr    error      // Variabel untuk menangkap error saat inisialisasi database
)

// initConfigDB menginisialisasi database konfigurasi menggunakan BadgerDB
func init() {
	// Membuka database Badger dengan opsi default
	dir, _ := ioutil.TempDir("", "badger-test")
	ConfigDB, dbErr = badger.Open(badger.DefaultOptions(dir).WithLogger(nil)) // Nonaktifkan logger untuk performa lebih baik
	if dbErr != nil {
		// Jika terjadi error, hentikan aplikasi dengan pesan error yang jelas
		log.Fatal("Gagal membuka database konfigurasi. Pesan error: " + dbErr.Error())
	}
}

// DatabaseConfiguration represents the configuration structure for a database
type DatabaseConfiguration struct {
	BlockReads  bool  `json:"block_reads"`  // Flag to block read operations
	BlockWrites bool  `json:"block_writes"` // Flag to block write operations
	SizeLimit   int64 `json:"size_limit"`   // Maximum size limit for the database in bytes
	Archived    bool  `json:"archived"`     // Flag to indicate if the database is archived
}

// Global variable to hold the default database configuration
var defaultDatabaseConfiguration *DatabaseConfiguration

// SaveDatabaseConfiguration saves or updates the configuration for a specific database
func SaveDatabaseConfiguration(databaseName string, mutex *sync.Mutex) error {
	// Serialize the configuration into JSON format
	configJSON, err := sonic.Marshal(defaultDatabaseConfiguration)
	if err != nil {
		// Return an error if serialization fails
		return err
	}

	// Start a new transaction for the metadata store
	transaction := ConfigDB.NewTransaction(true)

	// Lock the mutex to ensure thread safety
	mutex.Lock()
	defer mutex.Unlock()

	// Save the configuration in the metadata store with the database name as the key
	transaction.Set([]byte(databaseName), configJSON)

	// Commit the transaction to persist the changes
	if err := transaction.Commit(); err != nil {
		// Return an error if the commit fails
		return err
	}

	return nil // Return nil if everything succeeds
}

// UpdateDatabaseConfiguration updates the configuration for a specific database
func UpdateDatabaseConfiguration(databaseName string, newConfig *DatabaseConfiguration, mutex *sync.Mutex) error {
	// Serialize the new configuration into JSON format
	configJSON, err := sonic.Marshal(newConfig)
	if err != nil {
		// Return an error if serialization fails
		return err
	}

	// Start a new transaction for the metadata store
	transaction := ConfigDB.NewTransaction(true)

	// Lock the mutex to ensure thread safety
	mutex.Lock()
	defer mutex.Unlock()

	// Update the configuration in the metadata store with the database name as the key
	transaction.Set([]byte(databaseName), configJSON)

	// Commit the transaction to persist the changes
	if err := transaction.Commit(); err != nil {
		// Return an error if the commit fails
		return err
	}

	return nil // Return nil if the configuration is updated successfully
}

// DeleteDatabaseConfiguration removes the configuration for a specific database
func DeleteDatabaseConfiguration(databaseName string, mutex *sync.Mutex) error {
	// Start a new transaction for the metadata store
	transaction := ConfigDB.NewTransaction(true)

	// Lock the mutex to ensure thread safety
	mutex.Lock()
	defer mutex.Unlock()

	// Delete the configuration associated with the database name
	transaction.Delete([]byte(databaseName))

	// Commit the transaction to persist the changes
	if err := transaction.Commit(); err != nil {
		// Return an error if the commit fails
		return err
	}

	return nil // Return nil if the configuration is deleted successfully
}

// GetDatabaseConfiguration retrieves the configuration for a specific database
func GetDatabaseConfiguration(databaseName string) ([]byte, error) {
	// Start a new transaction for the metadata store
	transaction := ConfigDB.NewTransaction(false)

	// Retrieve the configuration associated with the database name
	item, err := transaction.Get([]byte(databaseName))
	if err != nil {
		// Return an error if retrieval fails
		return nil, err
	}

	// Extract the value from the result
	var configData []byte
	err = item.Value(func(val []byte) error {
		configData = val
		return nil
	})
	if err != nil {
		// Return an error if value extraction fails
		return nil, err
	}

	return configData, nil // Return the configuration data if everything succeeds
}
