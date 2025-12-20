package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Route("/freezers", func(r chi.Router) {
			r.Get("/", GetFreezers)
			r.Post("/", CreateFreezer)
			r.Delete("/{id}", DeleteFreezer)
		})

		r.Route("/items", func(r chi.Router) {
			r.Get("/", GetItems)
			r.Post("/", CreateItem)
			r.Post("/batch", CreateItemsBatch)
			r.Post("/consume", ConsumeItemsBatch)
			r.Post("/move", MoveItems)
			r.Put("/{id}", UpdateItem)
			r.Delete("/{id}", DeleteItem)
		})

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", GetCategories)
			r.Post("/", CreateCategory)
		})
	})

	return r
}
