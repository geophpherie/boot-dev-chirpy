package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jbeyer16/boot-dev-chirpy/internal/database"
)

func (cfg *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request) {
	type requestParameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}

	// process request body
	decoder := json.NewDecoder(r.Body)
	params := requestParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Print(err)
		respondWithError(w, http.StatusInternalServerError, "Invalid request body")
		return
	}

	// we only want upgrade events
	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusOK, "")
		return
	}

	// upgrade the user
	_, err = cfg.db.UpgradeUser(params.Data.UserId)
	if err != nil {
		if err == database.ErrUserNotFound {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Unable to upgrade user")
		return
	}

	respondWithJSON(w, http.StatusOK, "")
}
