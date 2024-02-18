package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jbeyer16/boot-dev-chirpy/internal"
)

var DB *internal.DB
var err error

func main() {
	DB, err = internal.NewDB("database.json")

	if err != nil {
		fmt.Println("Unable to read database")
		return
	}

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
	apiRouter.Post("/chirps", createChirp)
	apiRouter.Get("/chirps", getChirps)
	apiRouter.Get("/chirps/{chirpId}", getChirpById)
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)
	r.Mount("/admin", adminRouter)

	server := &http.Server{Addr: ":" + port, Handler: corsMux}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

func cleanseProfanity(msg string) (cleansedMsg string) {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	// convert to all lowercase
	words := strings.Split(msg, " ")
	// split on spaces
	for i, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	errResp := errorResponse{msg}

	dat, err := json.Marshal(errResp)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(code)
	w.Write(dat)
}
