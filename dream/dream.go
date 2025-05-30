package dream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/bedroompop/consist"
	"github.com/trianglehasfoursides/bedroompop/database"
	"github.com/trianglehasfoursides/bedroompop/flags"
	"github.com/trianglehasfoursides/bedroompop/pop"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var dbMutex sync.Mutex

func Start(ch chan os.Signal) {
	// router
	r := chi.NewRouter()

	// middlewares
	r.Use(MiddlewareAuth)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// routes
	r.Route("/v1", func(r chi.Router) {
		r.Route("/databases", func(r chi.Router) {
			r.Post("/", Create)
			r.Get("/{name}", Get)
			r.Delete("/{name}", Drop)
			r.Post("/query", Query)
			r.Post("/exec", Exec)
		})
	})

	// HTTP server
	server := &http.Server{
		Addr:    flags.HTTPAddr,
		Handler: r,
	}

	go func() {
		select {
		case _ = <-ch:
			server.Shutdown(context.TODO())
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.L().Sugar().Panic(err.Error())
	}
}

// CreateDatabase handles csreating a new database.
func Create(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	name, migration := gjson.Get(string(body), "name").String(), gjson.Get(string(body), "migration").String()
	if name == "" {
		http.Error(w, "database name is required", 400)
		return
	}
	command(w, "create", name, migration, "")
	return
}

// GetDatabase handles retrieving a specific database by name.
func Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "database name is required", 400)
		return
	}
	command(w, "get", name, "", "")
	return
}

// DeleteDatabase handles deleting a specific database by name.
func Drop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "database name is required", 400)
		return
	}
	command(w, "drop", name, "", "")
	return
}

func Query(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	name, query := gjson.Get(string(body), "name").String(), gjson.Get(string(body), "query").String()
	if name == "" && query == "" {
		http.Error(w, "database name is required", 400)
		return
	}
	command(w, "query", name, "", query)
	return
}

func Exec(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	name, query := gjson.Get(string(body), "name").String(), gjson.Get(string(body), "query").String()
	fmt.Println(query)
	if name == "" && query == "" {
		http.Error(w, "database name is required", 400)
		return
	}
	command(w, "exec", name, "", query)
	return
}

func command(w http.ResponseWriter, command string, dbname string, migration string, query string) {
	node := consist.Consist.LocateKey([]byte(dbname))
	if node.String() == flags.GRPCAddr {
		switch command {
		case "create":
			response(w, "sucess", database.Create(dbname, migration))
			return
		case "get":
			response(w, "sucess", database.Get(dbname))
			return
		case "drop":
			response(w, "sucess", database.Drop(dbname))
			return
		case "query":
			result, err := database.Query(dbname, query)
			if result == nil {
				write(w, http.StatusNotFound, map[string]string{"error": "empty"})
				return
			}
			response(w, result, err)
			return
		case "exec":
			result, err := database.Exec(dbname, query)
			response(w, result, err)
			return
		}
	}
	conn, err := grpc.NewClient(node.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		response(w, "error", err)
		return
	}
	defer conn.Close()
	client := pop.NewPopServiceClient(conn)
	switch command {
	case "create":
		status, err := client.Create(context.Background(), &pop.RequestCreate{Name: dbname, Migration: migration})
		if err != nil {
			response(w, "error", err)
			return
		}
		response(w, status.GetMsg(), nil)
		return
	case "get":
		status, err := client.Get(context.Background(), &pop.RequestGetDrop{Name: dbname})
		if err != nil {
			response(w, "error", err)
			return
		}
		response(w, status.GetMsg(), nil)
	case "drop":
		status, err := client.Drop(context.Background(), &pop.RequestGetDrop{Name: dbname})
		if err != nil {
			response(w, "error", err)
			return
		}
		response(w, status.GetMsg(), nil)
	case "query":
		result, err := client.Query(context.Background(), &pop.RequestQueryExec{Name: dbname, Query: query})
		if err != nil {
			response(w, "error", err)
			return
		}
		response(w, result.GetResult(), nil)
	case "exec":
		result, err := client.Exec(context.Background(), &pop.RequestQueryExec{Name: dbname, Query: query})
		if err != nil {
			response(w, "error", err)
			return
		}
		response(w, result.GetResult(), nil)
	}
}

func response(w http.ResponseWriter, result any, err error) {
	if err != nil {
		write(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	write(w, http.StatusOK, map[string]any{"message": result})
}

// write is a helper function to write JSON responses.
func write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
