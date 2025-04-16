package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path"

	"github.com/adrg/xdg"
	_ "github.com/marcboeker/go-duckdb/v2"
	"go.uber.org/zap"
)

// Query executes a SQL query on the specified SQLite database.
// It supports both SELECT queries (returns rows as JSON) and non-SELECT queries (logs affected rows).
func DuckQuery(databaseName string, query string, ctx context.Context) ([]byte, error) {
	// Open the SQLite database file
	db, err := sql.Open("duckdb", path.Join(xdg.DataHome, databaseName+".duck"))
	if err != nil {
		zap.L().Error("Failed to open database", zap.String("database", databaseName), zap.Error(err))
		return []byte(""), fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			zap.L().Error("Failed to close database connection", zap.String("database", databaseName), zap.Error(closeErr))
		}
	}()

	// Begin a transaction
	txn, err := db.Begin()
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.String("database", databaseName), zap.Error(err))
		return []byte(""), fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := txn.Rollback(); rollbackErr != nil {
			zap.L().Error("Failed to rollback transaction", zap.String("database", databaseName), zap.Error(rollbackErr))
		}
	}()

	// Execute the query and check if it returns rows
	rows, err := txn.QueryContext(ctx, query)
	if err != nil {
		// If the query does not return rows, execute it as a non-SELECT query
		result, execErr := txn.ExecContext(ctx, query)
		if execErr != nil {
			zap.L().Error("Failed to execute query", zap.String("database", databaseName), zap.String("query", query), zap.Error(execErr))
			return []byte(""), fmt.Errorf("failed to execute query: %w", execErr)
		}

		// Log the number of affected rows
		affectedRows, _ := result.RowsAffected()
		zap.L().Info("DML query executed successfully", zap.String("database", databaseName), zap.String("query", query), zap.Int64("affected_rows", affectedRows))
		return []byte(""), nil // Non-SELECT query executed successfully
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			zap.L().Error("Failed to close rows", zap.String("database", databaseName), zap.Error(closeErr))
		}
	}()

	// Retrieve column names from the result set
	columns, err := rows.Columns()
	if err != nil {
		zap.L().Error("Failed to retrieve columns", zap.String("database", databaseName), zap.Error(err))
		return []byte(""), fmt.Errorf("failed to retrieve columns: %w", err)
	}

	// Prepare a slice to hold the result rows
	var resultRows []map[string]interface{}

	// Iterate over the rows and process each one
	for rows.Next() {
		// Prepare slices to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		// Assign pointers to the values slice
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the current row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			zap.L().Error("Failed to scan row", zap.String("database", databaseName), zap.Error(err))
			return []byte(""), fmt.Errorf("failed to scan row: %w", err)
		}

		// Map the column names to their corresponding values
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
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

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		zap.L().Error("Error during row iteration", zap.String("database", databaseName), zap.Error(err))
		return []byte(""), fmt.Errorf("error during row iteration: %w", err)
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		zap.L().Error("Failed to commit transaction", zap.String("database", databaseName), zap.Error(err))
		return []byte(""), fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert the result rows to JSON
	jsonData, err := json.Marshal(resultRows)
	if err != nil {
		zap.L().Error("Failed to marshal rows to JSON", zap.String("database", databaseName), zap.Error(err))
		return []byte(""), fmt.Errorf("failed to marshal rows to JSON: %w", err)
	}

	zap.L().Info("SELECT query executed successfully", zap.String("database", databaseName), zap.String("query", query), zap.Int("rows_returned", len(resultRows)))
	return jsonData, nil // Return the JSON result
}
