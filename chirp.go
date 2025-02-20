package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/thetsajeet/chirpy/internal/helper"
)

func HandlerChipsValidator(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type validResponse struct {
		Cleaned_Body string `json:"cleaned_body"`
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	data, err := io.ReadAll(r.Body)
	if err != nil {
		helper.RespondWithError(w, 500, "Couldn't decode parameters", err)
		return
	}

	params := parameters{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		helper.RespondWithError(w, 500, "Couldn't unmarshall json", err)
		return
	}

	if len(params.Body) > 140 {
		helper.RespondWithError(w, 400, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	helper.RespondWithJson(w, 200, validResponse{
		Cleaned_Body: getCleanedBody(params.Body, badWords),
	})
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
