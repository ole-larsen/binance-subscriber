// Package router contains all routes for server. Based on chi router
package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/ole-larsen/binance-subscriber/internal/httpserver/handlers"
	"github.com/ole-larsen/binance-subscriber/internal/storage"
)

type Mux struct {
	Router  chi.Router
	storage storage.Storage
}

func NewMux() *Mux {
	return &Mux{
		Router: chi.NewRouter(),
	}
}

func (m *Mux) SetStorage(store storage.Storage) *Mux {
	m.storage = store
	return m
}

func (m *Mux) SetMiddlewares() *Mux {
	m.Router.Use(middleware.RequestID)
	m.Router.Use(middleware.RealIP)
	m.Router.Use(middleware.Recoverer)

	// disable logging because of ws
	// m.Router.Use(middlewares.LoggingMiddleware)

	return m
}

func (m *Mux) SetHandlers() *Mux {
	m.Router.Get("/ws", handlers.WebSocketHandler(m.storage))
	m.Router.Get("/status", handlers.StatusHandler)
	m.Router.Mount("/debug", middleware.Profiler())
	m.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"), // The url pointing to API definition
	))

	return m
}
