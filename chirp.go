package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thetsajeet/chirpy/internal/auth"
	"github.com/thetsajeet/chirpy/internal/database"
	"github.com/thetsajeet/chirpy/internal/helper"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helper.RespondWithError(w, 401, "unauthorized", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JWT_SECRET)
	if err != nil {
		helper.RespondWithError(w, 401, "unauthorized", err)
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

	cleanedBody := getCleanedBody(params.Body, badWords)

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userId,
	})

	if err != nil {
		helper.RespondWithError(w, 400, "Unable to create chirp", err)
		return
	}

	helper.RespondWithJson(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
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

func (cfg *apiConfig) AllChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	chirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		helper.RespondWithError(w, 400, "unable to get all chirps", err)
		return
	}

	resp := make([]Chirp, 0)

	for _, chirp := range chirps {
		resp = append(resp, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	helper.RespondWithJson(w, 200, resp)
}

func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		helper.RespondWithError(w, 400, "invalid chirpID", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpID)
	if err != nil {
		helper.RespondWithError(w, 404, "not found", err)
		return
	}

	helper.RespondWithJson(w, 200, Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	})
}

func (cfg *apiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helper.RespondWithError(w, 401, "not authenticated", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JWT_SECRET)
	if err != nil {
		helper.RespondWithError(w, 401, "not authenticated", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		helper.RespondWithError(w, 400, "invalid chirp id", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpID)
	if err != nil {
		helper.RespondWithError(w, 404, "chirp not found", err)
		return
	}

	if chirp.UserID != userId {
		helper.RespondWithError(w, 403, "unauthorized", err)
		return
	}

	if err := cfg.dbQueries.DeleteChirp(r.Context(), database.DeleteChirpParams{
		UserID: userId,
		ID:     chirpID,
	}); err != nil {
		helper.RespondWithError(w, 404, "unable to delete", err)
		return
	}

	helper.RespondWithJson(w, 204, map[string]any{})
}
