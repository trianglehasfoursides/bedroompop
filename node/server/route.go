package server

import "github.com/go-chi/chi/v5"

func route() *chi.Mux {
	router := chi.NewRouter()

	// index
	router.Get("/", index)

	// create
	router.Post("/databases", createDb)

	return router
}
