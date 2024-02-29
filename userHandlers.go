package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
		Email      string `json:"email"`
		Password   string `json:"password"`
		Expiration *int   `json:"expires_in_seconds,omitempty"`
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

	tokenDuration := math.Min(float64(*params.Expiration), 86400.0)
	dur, _ := time.ParseDuration(fmt.Sprintf("%vs", tokenDuration))

	issuedTime := jwt.NewNumericDate(time.Now().UTC())
	expiredTime := jwt.NewNumericDate(issuedTime.Add(dur))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  issuedTime,
			ExpiresAt: expiredTime,
			Subject:   fmt.Sprintf("%v", usr.Id),
		})

	signedToken, err := token.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		fmt.Print(err)
		respondWithError(w, http.StatusInternalServerError, "Unable to sign")
		return
	}
	authenticatedUser := struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}{Id: usr.Id, Email: usr.Email, Token: signedToken}

	respondWithJSON(w, http.StatusOK, authenticatedUser)
}
