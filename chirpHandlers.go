package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jbeyer16/boot-dev-chirpy/internal"
	"golang.org/x/exp/slices"
)

func createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	numChars := len(params.Body)

	if numChars > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedMessage := cleanseProfanity(params.Body)

	chirp, err := DB.CreateChirp(cleanedMessage)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating Chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func getChirps(w http.ResponseWriter, r *http.Request) {
	Chirps, err := DB.GetChirps()

	slices.SortFunc(Chirps, func(a, b internal.Chirp) int {
		return cmp.Compare(a.Id, b.Id)
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting Chirps")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirps)
}

func getChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId, err := strconv.Atoi(chi.URLParam(r, "chirpId"))
	if err != nil {
		fmt.Print("ERROR")
	}
	Chirp, err := DB.GetChirp(chirpId)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp)
}
