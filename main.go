package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jbeyer16/boot-dev-chirpy/internal/database"
	"github.com/joho/godotenv"
)

const port = "8080"

const databaseFile = "database.json"

func main() {
	godotenv.Load()

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		os.Remove(databaseFile)
	}

	DB, err := database.NewDB(databaseFile)
	if err != nil {
		fmt.Println("Unable to read database")
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	apiCfg := apiConfig{fileserverHits: 0, db: DB, jwtSecret: jwtSecret}

	r := chi.NewRouter()
	corsMux := middlewareCors(r)

	fsHandler := apiCfg.middlewareMetricsInc(fileServerHandler)
	r.Handle("/app/*", fsHandler)
	r.Handle("/app", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthHandler)
	apiRouter.Post("/chirps", apiCfg.createChirp)
	apiRouter.Get("/chirps", apiCfg.getChirps)
	apiRouter.Get("/chirps/{chirpId}", apiCfg.getChirpById)
	apiRouter.Post("/users", apiCfg.createUser)
	apiRouter.Post("/login", apiCfg.loginUser)
	apiRouter.Put("/users", apiCfg.updateUser)
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)
	adminRouter.Get("/reset", apiCfg.metricsResetHandler)
	r.Mount("/admin", adminRouter)

	server := &http.Server{Addr: ":" + port, Handler: corsMux}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
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
