package main

import (
	"github.com/alonzzio/log-monitoring-server/internal/access"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// routes use and implement all routes
func routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)

	mux.Get("/ping", access.Repo.ServerPing)
	mux.Get("/service-severity-stat", access.Repo.GetServiceSeverity)
	return mux
}
