package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/thetsajeet/chirpy/internal/database"
	"github.com/thetsajeet/chirpy/internal/helper"
)

func main() {
	godotenv.Load()

	dbUrl := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		dbQueries:      database.New(db),
		PLATFORM:       os.Getenv("PLATFORM"),
		JWT_SECRET:     os.Getenv("JWT_SECRET"),
	}

	if apiCfg.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	filepathHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInfo(filepathHandler))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerFileServerHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetrics)
	mux.HandleFunc("GET /api/chirps", apiCfg.AllChirps)
	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirp)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirp)

	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerFileServerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	htmlTemplate := `
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
	`
	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(htmlTemplate, cfg.fileServerHits.Load())))
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	if cfg.PLATFORM != "dev" {
		w.WriteHeader(403)
	}

	cfg.fileServerHits.Store(0)
	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		helper.RespondWithError(w, 400, "Couldn't delete users", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
