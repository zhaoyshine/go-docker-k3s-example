package api

import (
	"fmt"
	"log"
	"net/http"

	"k3sdemo/internal/config"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func StartHttp(cfg *config.Config) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	log.Println("http start")

	err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.HTTP.Port), r)
	if err != nil {
		log.Printf("http start failed: %v\n", err)
	}
}
