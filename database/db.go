package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var sqlite = ".sqlite"

func Create(databaseName string, migration string) error {
	mtx := new(sync.Mutex)
	mtx.Lock()
	defer mtx.Unlock()

	// Check if the database file already exists
	if _, err := os.Stat(databaseName + sqlite); err == nil {
		return errors.New("database already exist")
	}

	// Create SQLite database
	if _, err := os.Create(databaseName + sqlite); err != nil {
		return err
	}

	if migration != "" {

		db, err := sql.Open("sqlite3", databaseName+sqlite)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), migration); err != nil {
			return errors.New("can't run migration due to error : " + err.Error())
		}
	}

	return nil //  created successfully
}

// Delete deletes a database file (SQLite, BoltDB, or DuckDB) along with its configuration.
func Drop(databaseName string) error {
	// Construct the full path for the database file
	databasePath := databaseName + sqlite

	// Attempt to remove the database file
	_ = os.Remove(databasePath)
	return nil
}

// Get retrieves the configuration for a specific database.
func Get(databaseName string) error {
	// Construct the full path for the database file
	databasePath := databaseName + sqlite

	// Check if the database file exists
	if _, err := os.Stat(databasePath); err != nil {
		return errors.New("database not found")
	}

	return nil
}

// Query executes a SQL query on the specified SQLite database.
// It supports both SELECT queries (returns rows as JSON) and non-SELECT queries (logs affected rows).
func Query(databaseName string, query string) ([]byte, error) {
	if err := Get(databaseName); err != nil {
		return nil, err
	}
	ctx := context.Background()
	// Open the SQLite database file
	db, err := sql.Open("sqlite3", databaseName+sqlite)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Begin a transaction
	txn, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	// Execute the query
	rows, err := txn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// Retrieve column names from the result set
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare a slice to hold the result rows
	var resultRows []map[string]any

	// Iterate over the rows and process each one
	for rows.Next() {
		// Prepare slices to hold column values
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		// Assign pointers to the values slice
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the current row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Map the column names to their corresponding values
		rowMap := make(map[string]any)
		for i, col := range columns {
			var v any
			val := values[i]

			// Convert []byte to string for readability
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}

			rowMap[col] = v
		}

		// Append the row to the result set
		resultRows = append(resultRows, rowMap)
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		return nil, err
	}

	// Convert the result rows to JSON
	jsonData, _ := json.Marshal(resultRows)
	return jsonData, nil // Return the JSON result
}

// Exec executes a non-SELECT SQL query (e.g., INSERT, UPDATE, DELETE) on the specified SQLite database.
// It returns the number of rows affected.
func Exec(databaseName string, query string) (int64, error) {
	if err := Get(databaseName); err != nil {
		return 0, err
	}
	ctx := context.Background()
	// Open the SQLite database file
	db, err := sql.Open("sqlite3", databaseName+sqlite)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Begin a transaction
	txn, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer txn.Rollback()

	// Execute the query
	result, err := txn.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		return 0, err
	}

	// Get the number of rows affected
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil // Return the number of rows affected
}
