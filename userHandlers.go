package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	usr, err := cfg.db.CreateUser(params.Email, params.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating User")
		return
	}

	userNoPassword := struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}{Id: usr.Id, Email: usr.Email}

	respondWithJSON(w, http.StatusCreated, userNoPassword)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	// get password from request
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// get user with this email
	usr, err := cfg.db.GetUser(params.Email)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(usr.HashedPassword), []byte(params.Password))

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	userNoPassword := struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}{Id: usr.Id, Email: usr.Email}

	respondWithJSON(w, http.StatusOK, userNoPassword)
}
