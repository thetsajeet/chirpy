package main

import (
	"sync/atomic"

	"github.com/thetsajeet/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
	PLATFORM       string
	JWT_SECRET     string
	POLKA_KEY      string
}
