package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jbeyer16/boot-dev-chirpy/internal/auth"
	"github.com/jbeyer16/boot-dev-chirpy/internal/database"
)

// define user so that password won't be written to json
type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type requestParameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// process request body
	decoder := json.NewDecoder(r.Body)
	params := requestParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid request body")
		return
	}

	// hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	// create user
	user, err := cfg.db.CreateUser(params.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, database.ErrUserAlreadyExists) {
			respondWithError(w, http.StatusConflict, "User already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error creating User")
		return
	}

	// send back user
	response := User{
		Id:    user.Id,
		Email: user.Email,
	}
	respondWithJSON(w, http.StatusCreated, response)
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request")
		return
	}

	// get user with this email
	user, err := cfg.db.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find user")
		return
	}

	// validate password with hashed password
	err = auth.ValidatePassword(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	// issue the access token
	accessToken, err := auth.IssueJWT("chirpy-access", user.Id, cfg.jwtSecret, time.Duration(1)*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to issue access token")
		return
	}

	// issue the refresh token
	refreshToken, err := auth.IssueJWT("chirpy-refresh", user.Id, cfg.jwtSecret, time.Duration(24*60)*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to issue refresh token")
		return
	}

	// add refresh token to database
	err = cfg.db.AddToken(refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to issue refresh token")
		return
	}

	// return authenticated user response
	authenticatedUser := struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		Id:           user.Id,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, authenticatedUser)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
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

	// parse the request body
	type requestParameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := requestParameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid request body")
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

	// hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	// update the user in the database
	updatedUser, err := cfg.db.UpdateUser(userIdNum, params.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to update user")
		return
	}

	// return updated user info
	response := User{
		Id:    updatedUser.Id,
		Email: updatedUser.Email,
	}
	respondWithJSON(w, http.StatusOK, response)
}
