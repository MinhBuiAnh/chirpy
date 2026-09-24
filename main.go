package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/MinhBuiAnh/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Get the environment variable
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")

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
		platform: platform,
		secret: secret,
	}
	
	mux.Handle("/app", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", handleReadiness)
	
	mux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", cfg.handleUpdateEmailPassword)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)

	mux.HandleFunc("POST /api/refresh", cfg.handleRefreshToken)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevokeToken)

	mux.HandleFunc("POST /api/chirps", cfg.handleCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", cfg.handleGetChirpById)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handleDeleteChirpById)

	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handleReset)

	server := &http.Server{
		Handler: mux,
		Addr: ":" + port,
	}
	
	log.Printf("Starting server on port %s, serving files from %s", port, filepathRoot)
	log.Fatal(server.ListenAndServe())
}