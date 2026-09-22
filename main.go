package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/MinhBuiAnh/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Get the environment variable
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")

	// Connect to the database
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Cannot connect to the database")
	}

	const filepathRoot = "."
	const port = "8080"

	// Configure and setting up the server
	mux := http.NewServeMux()
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db: database.New(db),
	}
	
	mux.Handle("/app", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handleReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)

	server := &http.Server{
		Handler: mux,
		Addr: ":" + port,
	}
	
	log.Printf("Starting server on port %s, serving files from %s", port, filepathRoot)
	log.Fatal(server.ListenAndServe())
}

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqParams{}

	w.Header().Set("Content-Type", "application/json")

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedBody := replaceBadWords(params.Body)
	resBody := returnVals{
		CleanedBody: cleanedBody,
	}
	respondWithJSON(w, http.StatusOK, resBody)
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	content := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	w.Write([]byte(content))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}