package database

import (
	"path"

	"github.com/adrg/xdg"
	"go.etcd.io/bbolt"
)

const defaultBucket = "defaultBucket" // Define the default bucket name

// OpenBoltDatabase opens a BoltDB database file with the given name.
// It returns a pointer to the database instance or an error if the operation fails.
func OpenBoltDatabase(databaseName string) (*bbolt.DB, error) {
	// Construct the full path for the BoltDB file
	databasePath := path.Join(xdg.DataHome, databaseName+".bolt")

	// Open the BoltDB file with default options
	db, err := bbolt.Open(databasePath, 0600, bbolt.DefaultOptions)
	if err != nil {
		return nil, err
	}

	// Ensure the default bucket exists
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(defaultBucket))
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// GetValue retrieves a value from the default bucket in a BoltDB database for the specified key.
// The retrieved value is returned as a byte slice or an error if the operation fails.
func GetValue(databaseName string, key []byte) ([]byte, error) {
	// Open the BoltDB database
	db, err := OpenBoltDatabase(databaseName)
	if err != nil {
		return nil, err
	}
	defer db.Close() // Ensure the database is closed after the operation

	var value []byte

	// Start a read-only transaction
	err = db.View(func(tx *bbolt.Tx) error {
		// Retrieve the default bucket
		bucket := tx.Bucket([]byte(defaultBucket))

		// Get the value for the specified key
		value = bucket.Get(key)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// PutValue stores a key-value pair in the default bucket of a BoltDB database.
// It requires the database name, key, and value as input.
// Returns an error if the operation fails.
func PutValue(databaseName string, key []byte, value []byte) error {
	// Open the BoltDB database
	db, err := OpenBoltDatabase(databaseName)
	if err != nil {
		return err
	}
	defer db.Close() // Ensure the database is closed after the operation

	// Start a read-write transaction
	err = db.Update(func(tx *bbolt.Tx) error {
		// Retrieve the default bucket
		bucket := tx.Bucket([]byte(defaultBucket))

		// Put the key-value pair into the bucket
		return bucket.Put(key, value)
	})

	return err
}

// DeleteValue removes a key-value pair from the default bucket in a BoltDB database.
// It requires the database name and key as input.
// Returns an error if the operation fails.
func DeleteValue(databaseName string, key []byte) error {
	// Open the BoltDB database
	db, err := OpenBoltDatabase(databaseName)
	if err != nil {
		return err
	}
	defer db.Close() // Ensure the database is closed after the operation

	// Start a read-write transaction
	err = db.Update(func(tx *bbolt.Tx) error {
		// Retrieve the default bucket
		bucket := tx.Bucket([]byte(defaultBucket))

		// Delete the key-value pair from the bucket
		return bucket.Delete(key)
	})

	return err
}
