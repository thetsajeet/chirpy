package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}

	filepathHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInfo(filepathHandler))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerFileServerHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerChipsValidator)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileServerHits atomic.Int32
}

func handlerChipsValidator(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type errorResonse struct {
		Error string `json:"error"`
	}

	type validResponse struct {
		Cleaned_Body string `json:"cleaned_body"`
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	params := parameters{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	if len(params.Body) > 140 {
		errorResp := errorResonse{
			Error: "Chirp is too long",
		}
		respondWithJSON(w, 400, errorResp)
		return
	}

	clean_body := cleanBody(params.Body)

	successResp := validResponse{
		Cleaned_Body: clean_body,
	}

	respondWithJSON(w, 200, successResp)
}

func cleanBody(body string) string {
	bodySl := strings.Fields(body)
	resSl := make([]string, 0)
	invalidWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, b := range bodySl {
		for _, inv := range invalidWords {
			if strings.Compare(strings.ToLower(b), inv) == 0 {
				b = "****"
				break
			}
		}
		resSl = append(resSl, b)
	}
	res := strings.Join(resSl, " ")
	return res
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
	cfg.fileServerHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
