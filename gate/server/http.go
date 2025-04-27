package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/trianglehasfoursides/mathrock/node/database"
	"go.uber.org/zap"
)

var dbMutex sync.Mutex // Mutex untuk operasi thread-safe pada database

// StartHTTPServer initializes and starts the HTTP server with Chi router.
func StartHTTPServer(ctx context.Context, addr string) error {
	// Create a new Chi router
	r := chi.NewRouter()

	// Add middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second)) // Set a timeout for all requests

	// Define routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/databases", func(r chi.Router) {
			r.Get("/", listDatabases)           // GET /api/v1/databases
			r.Post("/", createDatabase)         // POST /api/v1/databases
			r.Get("/{name}", getDatabase)       // GET /api/v1/databases/{name}
			r.Delete("/{name}", deleteDatabase) // DELETE /api/v1/databases/{name}
		})
	})

	// Create the HTTP server
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Graceful shutdown handling
	go func() {
		<-ctx.Done()
		zap.L().Info("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("Failed to gracefully shut down HTTP server", zap.Error(err))
		}
	}()

	// Start the server
	zap.L().Info("Starting HTTP server...", zap.String("address", addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.L().Error("HTTP server failed", zap.Error(err))
		return err
	}

	return nil
}

// ListDatabases handles listing all databases.
func listDatabases(w http.ResponseWriter, r *http.Request) {
	pipe := make(chan *Message)
	pipe <- &Message{
		Key: "list_db",
	}
	msg := new(ChanMessage)
	msg.Pipe = pipe
	chanMessage <- msg
}

// CreateDatabase handles creating a new database.
func createDatabase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := database.CreateDatabase(req.Name, &dbMutex); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Database created successfully", "name": req.Name})
}

// GetDatabase handles retrieving a specific database by name.
func getDatabase(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	config, err := database.GetDatabase(name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"name": name, "configuration": json.RawMessage(config)})
}

// DeleteDatabase handles deleting a specific database by name.
func deleteDatabase(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := database.DeleteDatabase(name, &dbMutex); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Database deleted successfully", "name": name})
}

// writeJSON is a helper function to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
