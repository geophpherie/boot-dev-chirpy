package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jbeyer16/boot-dev-chirpy/internal/auth"
)

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	// check for empty body only
	if r.Body != http.NoBody {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// check for proper bearer token
	token, err := auth.ParseBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid token")
		return
	}

	// ensure token is valid
	_, claims, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// ensure token is a refresh token
	issuer, err := claims.GetIssuer()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to parse token")
		return
	}
	if issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Non-refresh token received.")
		return
	}

	// ensure token has not been revoked
	err = cfg.db.CheckToken(token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token has been revoked!")
		return
	}

	// create new access token
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
	accessToken, err := auth.IssueJWT("chirpy-access", userIdNum, cfg.jwtSecret, time.Duration(1)*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to issue access token")
		return
	}

	// return new token
	response := struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	// check for empty body only
	if r.Body != http.NoBody {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// check for proper bearer token
	token, err := auth.ParseBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid token")
		return
	}

	// ensure token is valid
	_, claims, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// ensure token is a refresh token
	issuer, err := claims.GetIssuer()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to parse token")
		return
	}
	if issuer != "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Non-refresh token received.")
		return
	}

	err = cfg.db.RevokeToken(token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to revoke access token")
	}

	respondWithJSON(w, http.StatusOK, "")
}
