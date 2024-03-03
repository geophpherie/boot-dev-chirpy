package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jbeyer16/boot-dev-chirpy/internal/auth"
	"github.com/jbeyer16/boot-dev-chirpy/internal/database"
	"golang.org/x/exp/slices"
)

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.ParseBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid token")
		return
	}

	_, claims, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	issuer, err := claims.GetIssuer()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to parse token")
		return
	}

	if issuer != "chirpy-access" {
		respondWithError(w, http.StatusUnauthorized, "Non-access token received.")
		return
	}
	// convert id from string to int
	userId, err := claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to parse user Id")
		return
	}
	userIdNum, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to parse user Id")
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)

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

	chirp, err := cfg.db.CreateChirp(cleanedMessage, userIdNum)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating Chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
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

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	Chirps, err := cfg.db.GetChirps()

	slices.SortFunc(Chirps, func(a, b database.Chirp) int {
		return cmp.Compare(a.Id, b.Id)
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting Chirps")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirps)
}

func (cfg *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId, err := strconv.Atoi(chi.URLParam(r, "chirpId"))
	if err != nil {
		fmt.Print("ERROR")
	}
	Chirp, err := cfg.db.GetChirp(chirpId)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "")
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp)
}
