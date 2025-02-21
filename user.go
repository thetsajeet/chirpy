package main

import (
	"encoding/json"
	"fmt"
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
	Token     string    `json:"token,omitempty"`
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

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := auth.MakeJWT(dat.ID, cfg.JWT_SECRET)
	if err != nil {
		helper.RespondWithError(w, 500, "unable to create token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		helper.RespondWithError(w, 500, "unable to create refresh token", err)
		return
	}

	err = cfg.dbQueries.StoreRefreshToken(r.Context(), database.StoreRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dat.ID,
		ExpiresAt: time.Now().Add(60 * time.Hour * 24),
	})
	if err != nil {
		helper.RespondWithError(w, 400, "unable to create refresh token", err)
		return
	}

	helper.RespondWithJson(w, 200, response{
		User: User{
			ID:        dat.ID,
			Email:     dat.Email,
			CreatedAt: dat.CreatedAt,
			UpdatedAt: dat.UpdatedAt,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helper.RespondWithError(w, 401, "unauthorized", err)
		return
	}

	dat, err := cfg.dbQueries.LookupToken(r.Context(), refreshToken)
	if err != nil || dat.ExpiresAt.Compare(time.Now()) <= 0 || (dat.RevokedAt.Valid && dat.RevokedAt.Time.Compare(time.Now()) <= 0) {
		helper.RespondWithError(w, 401, "token expired or not found", err)
		return
	}

	token, err := auth.MakeJWT(dat.UserID, cfg.JWT_SECRET)
	if err != nil {
		helper.RespondWithError(w, 500, "unable to make JWT", err)
		return
	}

	fmt.Printf("%v", token)

	helper.RespondWithJson(w, 200, map[string]any{
		"token": token,
	})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helper.RespondWithError(w, 401, "unauthorized", err)
		return
	}

	if err := cfg.dbQueries.RevokeToken(r.Context(), refreshToken); err != nil {
		helper.RespondWithError(w, 401, "something went wrong", err)
		return
	}

	helper.RespondWithJson(w, 204, map[string]any{})
}

func (cfg *apiConfig) handleUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		helper.RespondWithError(w, 500, "unable to decode json", err)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		helper.RespondWithError(w, 401, "unauthorized", err)
		return
	}

	userId, err := auth.ValidateJWT(accessToken, cfg.JWT_SECRET)
	if err != nil {
		helper.RespondWithError(w, 401, "unauthorized", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		helper.RespondWithError(w, 401, "unable to hash password", err)
		return
	}

	dat, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userId,
	})
	if err != nil {
		helper.RespondWithError(w, 400, "unable to save to db", err)
		return
	}

	helper.RespondWithJson(w, 200, User{
		ID:        userId,
		CreatedAt: dat.CreatedAt,
		UpdatedAt: dat.UpdatedAt,
		Email:     dat.Email,
	})
}
