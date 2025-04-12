package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/trianglehasfoursides/mathrock/node/sqlite"
	"go.uber.org/zap"
)

// StartHTTP2Server starts an HTTP/2 server with Chi and TLS
func StartHTTP2Server(address, certFile, keyFile string) error {
	// Load the TLS certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate and key: %v", err)
	}

	// Configure the TLS settings
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// Create a Chi router
	router := chi.NewRouter()

	// Define RESTful routes
	router.Get("/databases", listDatabasesHandler)                      // List all databases
	router.Post("/databases", createDatabaseHandler)                    // Create a new database
	router.Get("/databases/{name}", getDatabaseHandler)                 // Retrieve a database
	router.Delete("/databases/{name}", deleteDatabaseHandler)           // Delete a database
	router.Put("/databases/{name}/config", updateDatabaseConfigHandler) // Update database configuration

	// Create an HTTP server with TLS
	server := &http.Server{
		Addr:      address,
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	zap.L().Info("HTTP/2 server started", zap.String("address", address))
	return server.ListenAndServeTLS("", "")
}

// createDatabaseHandler handles the creation of a new database
func createDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Database string `json:"database"`
	}

	// Parse the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		zap.L().Error("Invalid JSON format", zap.Error(err))
		http.Error(w, `{"error": "Invalid JSON format"}`, http.StatusBadRequest)
		return
	}

	// Validate the database name
	if request.Database == "" {
		zap.L().Error("Missing 'database' field")
		http.Error(w, `{"error": "Missing 'database' field"}`, http.StatusBadRequest)
		return
	}

	// Create the database
	mutex := &sync.Mutex{}
	if err := sqlite.InitializeDatabase(request.Database, mutex); err != nil {
		zap.L().Error("Error creating database", zap.String("database", request.Database), zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error creating database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	zap.L().Info("Database created successfully", zap.String("database", request.Database))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"success": "Database '%s' created successfully"}`, request.Database)))
}

// getDatabaseHandler handles retrieving data from a database
func getDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	databaseName := chi.URLParam(r, "name")

	// Validate the database name
	if databaseName == "" {
		zap.L().Error("Missing 'database' parameter")
		http.Error(w, `{"error": "Missing 'database' parameter"}`, http.StatusBadRequest)
		return
	}

	// Retrieve the database
	data, err := sqlite.RetrieveDatabase(databaseName)
	if err != nil {
		zap.L().Error("Error retrieving database", zap.String("database", databaseName), zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error retrieving database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with the data
	zap.L().Info("Database retrieved successfully", zap.String("database", databaseName))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success": "Data from database '%s'", "data": %s}`, databaseName, string(data))))
}

// deleteDatabaseHandler handles deleting a database
func deleteDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	databaseName := chi.URLParam(r, "name")

	// Validate the database name
	if databaseName == "" {
		zap.L().Error("Missing 'database' parameter")
		http.Error(w, `{"error": "Missing 'database' parameter"}`, http.StatusBadRequest)
		return
	}

	// Delete the database
	mutex := &sync.Mutex{}
	if err := sqlite.RemoveDatabase(databaseName, mutex); err != nil {
		zap.L().Error("Error deleting database", zap.String("database", databaseName), zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error deleting database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	zap.L().Info("Database deleted successfully", zap.String("database", databaseName))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success": "Database '%s' deleted successfully"}`, databaseName)))
}

// updateDatabaseConfigHandler handles updating the configuration of a database
func updateDatabaseConfigHandler(w http.ResponseWriter, r *http.Request) {
	databaseName := chi.URLParam(r, "name")

	// Validate the database name
	if databaseName == "" {
		zap.L().Error("Missing 'database' parameter")
		http.Error(w, `{"error": "Missing 'database' parameter"}`, http.StatusBadRequest)
		return
	}

	// Parse the JSON request body for the new configuration
	var newConfig sqlite.DatabaseConfiguration
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		zap.L().Error("Invalid JSON format", zap.Error(err))
		http.Error(w, `{"error": "Invalid JSON format"}`, http.StatusBadRequest)
		return
	}

	// Validate the configuration fields (optional, based on your requirements)
	if newConfig.SizeLimit < 0 {
		zap.L().Error("Invalid 'size_limit' value")
		http.Error(w, `{"error": "Invalid 'size_limit' value"}`, http.StatusBadRequest)
		return
	}

	// Update the database configuration
	mutex := &sync.Mutex{}
	if err := sqlite.UpdateDatabaseConfiguration(databaseName, &newConfig, mutex); err != nil {
		zap.L().Error("Error updating database configuration", zap.String("database", databaseName), zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error updating database configuration: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	zap.L().Info("Database configuration updated successfully", zap.String("database", databaseName))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success": "Configuration for database '%s' updated successfully"}`, databaseName)))
}

// listDatabasesHandler handles listing all databases
func listDatabasesHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the list of databases
	databases, err := sqlite.ListDatabases()
	if err != nil {
		zap.L().Error("Error listing databases", zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error listing databases: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with the list of databases
	zap.L().Info("Databases listed successfully", zap.Int("count", len(databases)))
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"success":   true,
		"databases": databases,
	}
	json.NewEncoder(w).Encode(response)
}
