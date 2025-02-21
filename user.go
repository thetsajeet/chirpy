package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thetsajeet/chirpy/internal/auth"
	"github.com/thetsajeet/chirpy/internal/database"
	"github.com/thetsajeet/chirpy/internal/helper"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		helper.RespondWithError(w, 500, "unable to hash password", err)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		helper.RespondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	helper.RespondWithJson(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	p := params{}
	err := decoder.Decode(&p)
	if err != nil {
		helper.RespondWithError(w, 500, "unable to unmarshall json", err)
		return
	}

	dat, err := cfg.dbQueries.LoginUser(r.Context(), p.Email)
	if err != nil {
		helper.RespondWithError(w, 401, "Unauthorized", err)
		return
	}

	if err = auth.CheckPasswordHash(dat.HashedPassword, p.Password); err != nil {
		helper.RespondWithError(w, 401, "Unauthorized", err)
		return
	}

	helper.RespondWithJson(w, 200, User{
		ID:        dat.ID,
		Email:     dat.Email,
		CreatedAt: dat.CreatedAt,
		UpdatedAt: dat.UpdatedAt,
	})
}
