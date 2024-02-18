package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	const port = "8080"

	apiCfg := apiConfig{fileserverHits: 0}

	r := chi.NewRouter()
	corsMux := middlewareCors(r)

	fsHandler := apiCfg.middlewareMetricsInc(fileServerHandler)
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthHandler)
	apiRouter.Get("/reset", apiCfg.metricsResetHandler)
	apiRouter.Post("/validate_chirp", validateChirpHandler)
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)

	r.Mount("/admin", adminRouter)

	server := &http.Server{Addr: ":" + port, Handler: corsMux}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
