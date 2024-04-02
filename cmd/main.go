package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"test/internal/api/handler/get_value"
	"test/internal/api/handler/save_value"
	myMiddleware "test/internal/api/middleware"
	"test/internal/polymorphism/storage/map_storage"
)

func main() {
	r := chi.NewRouter()
	// timer -> logger -> recoverer -> jsonHeader -> handler

	r.Use(myMiddleware.Timer)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(myMiddleware.JsonHeader)

	storage := map_storage.NewStorage()

	savePairHandler := save_value.NewHandler(storage)
	getValueHandler := get_value.NewHandler(storage)

	r.Method(http.MethodPost, "/save", savePairHandler)
	r.Method(http.MethodGet, "/find", getValueHandler)

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		panic("cannot create server")
	}
}
