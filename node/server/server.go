package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/quic-go/quic-go/http3"
	"github.com/trianglehasfoursides/mathrock/node/database"
	"go.uber.org/zap"
)

// validate is a validator instance for validating request data.
var validate = validator.New()

// ValidateMyVal implements validator.Func
func validateCategory(fl validator.FieldLevel) bool {
	return fl.Field().String() == "all" || fl.Field().String() == "sqlite" || fl.Field().String() == "bolt"
}

func init() {
	// Register custom validation for the "category" field
	validate.RegisterValidation("category", validateCategory)
}

// StartHTTP2Server starts an HTTP/2 server with Chi and HTTP/3 support.
func StartHTTP2Server(address, certFile, keyFile string) error {
	// Create a Chi router
	router := chi.NewRouter()

	// Define RESTful routes
	router.Route("/databases", func(r chi.Router) {
		router.Get("{category}", listDatabasesHandler)                   // List all databases
		router.Post("/", createDatabaseHandler)                          // Create a new database
		router.Get("/{name}", getDatabaseHandler)                        // Retrieve a database
		router.Delete("/", deleteDatabaseHandler)                        // Delete a database
		router.Put("/{name}/configuration", updateDatabaseConfigHandler) // Update database configuration
	})
	// Create an HTTP/3 server
	server := &http3.Server{
		Addr:    address,
		Handler: router,
	}

	// Log server start and listen for connections
	zap.L().Info("HTTP/2 server started", zap.String("address", address))
	return server.ListenAndServe()
}

// createDatabaseHandler handles the creation of a new database.
func createDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name" validate:"required"`
		Category string `json:"category" validate:"required"`
	}

	// Parse the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		zap.L().Error("Invalid JSON format", zap.Error(err))
		http.Error(w, `{"error": "Invalid JSON format"}`, http.StatusBadRequest)
		return
	}

	// Validate the database name and category
	if err := validate.Struct(request); err != nil {
		zap.L().Error("Validation error", zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Validation error: %v"}`, err), http.StatusBadRequest)
		return
	}

	// Create the database
	mutex := &sync.Mutex{}
	if err := database.CreateDatabase(request.Name, request.Category, mutex); err != nil {
		zap.L().Error("Error creating database", zap.String("database", request.Name), zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error creating database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	zap.L().Info("Database created successfully", zap.String("database", request.Name))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"success": "Database '%s' created successfully"}`, request.Name)))
}

// getDatabaseHandler handles retrieving data from a database.
func getDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	databaseName := chi.URLParam(r, "name")

	// Validate the database name
	if databaseName == "" {
		zap.L().Error("Missing 'database' parameter")
		http.Error(w, `{"error": "Missing 'database' parameter"}`, http.StatusBadRequest)
		return
	}

	// Retrieve the database
	data, err := database.GetDatabase(databaseName, "")
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

// deleteDatabaseHandler handles deleting a database.
func deleteDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name     string `json:"name" validate:"required"`
		Category string `json:"category" validate:"required"`
	}

	// Parse the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		zap.L().Error("Invalid JSON format", zap.Error(err))
		http.Error(w, `{"error": "Invalid JSON format"}`, http.StatusBadRequest)
		return
	}

	// Validate the database name and category
	if err := validate.Struct(request); err != nil {
		zap.L().Error("Validation error", zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Validation error: %v"}`, err), http.StatusBadRequest)
		return
	}

	// Delete the database
	mutex := &sync.Mutex{}
	if err := database.DeleteDatabase(request.Name, request.Category, mutex); err != nil {
		zap.L().Error("Error deleting database", zap.String("database", request.Name), zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error deleting database: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	zap.L().Info("Database deleted successfully", zap.String("database", request.Name))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success": "Database '%s' deleted successfully"}`, request.Name)))
}

// updateDatabaseConfigHandler handles updating the configuration of a database.
func updateDatabaseConfigHandler(w http.ResponseWriter, r *http.Request) {
	databaseName := chi.URLParam(r, "name")

	// Validate the database name
	if databaseName == "" {
		zap.L().Error("Missing 'database' parameter")
		http.Error(w, `{"error": "Missing 'database' parameter"}`, http.StatusBadRequest)
		return
	}

	// Parse the JSON request body for the new configuration
	var newConfig database.DatabaseConfiguration
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		zap.L().Error("Invalid JSON format", zap.Error(err))
		http.Error(w, `{"error": "Invalid JSON format"}`, http.StatusBadRequest)
		return
	}

	// Validate the configuration fields
	if newConfig.SizeLimit < 0 {
		zap.L().Error("Invalid 'size_limit' value")
		http.Error(w, `{"error": "Invalid 'size_limit' value"}`, http.StatusBadRequest)
		return
	}

	// Update the database configuration
	mutex := &sync.Mutex{}
	if err := database.UpdateDatabaseConfiguration(databaseName, "", &newConfig, mutex); err != nil {
		zap.L().Error("Error updating database configuration", zap.String("database", databaseName), zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error updating database configuration: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	zap.L().Info("Database configuration updated successfully", zap.String("database", databaseName))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"success": "Configuration for database '%s' updated successfully"}`, databaseName)))
}

// listDatabasesHandler handles listing all databases.
func listDatabasesHandler(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")

	// validate the category parameter
	if err := validate.Var(category, "category"); err != nil {
		zap.L().Error("Invalid 'category' parameter", zap.Error(err))
		http.Error(w, `{"error": "Invalid 'category' parameter"}`, http.StatusBadRequest)
		return
	}

	// Retrieve the list of databases based on the category
	databases, err := database.ListDatabases(category)
	if err != nil {
		zap.L().Error("Error listing databases", zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error": "Error listing databases: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Respond with the list of databases
	zap.L().Info("Databases listed successfully", zap.Int("count", len(databases)))
	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"success":   true,
		"databases": databases,
	}
	json.NewEncoder(w).Encode(response)
}
