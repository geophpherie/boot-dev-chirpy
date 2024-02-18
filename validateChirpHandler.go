package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type error struct {
		Error string `json:"error"`
	}

	type valid struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respBody := error{Error: "Something went wrong"}

		dat, _ := json.Marshal(respBody)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(dat)
		return
	}

	numChars := len(params.Body)

	fmt.Println(numChars)
	if numChars > 140 {
		respBody := error{Error: "Chirp is too long"}

		dat, _ := json.Marshal(respBody)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(dat)
		return
	}

	respBody := valid{Valid: true}

	dat, _ := json.Marshal(respBody)

	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}
