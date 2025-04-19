package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/SergioFloresCorrea/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func main() {
	const port = "8080"
	err := godotenv.Load()
	if err != nil {
		log.Printf("%v\n", err)
		os.Exit(1)
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("We couldn't access the database: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	apiCfg := &apiConfig{dbQueries: dbQueries, platform: platform}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", ServerReady)
	mux.HandleFunc("GET /admin/metrics", apiCfg.CountRequests)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetCounterRequests)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.ValidateAndSaveChirp)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}
