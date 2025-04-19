package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) ResetCounterRequests(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	prevCount := cfg.fileserverHits.Swap(0)
	fmt.Fprintf(w, "The number of hits has been reset from %d to %d.\n", prevCount, cfg.fileserverHits.Load())

	cfg.dbQueries.DeleteUsers(req.Context())
	fmt.Fprintf(w, "All users data has been deleted.\n")

}

func (cfg *apiConfig) CountRequests(w http.ResponseWriter, req *http.Request) {
	htmlForm := `
	<html>
     <body>
      <h1>Welcome, Chirpy Admin</h1>
      <p>Chirpy has been visited %d times!</p>
     </body>
    </html>
	`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, htmlForm, cfg.fileserverHits.Load())
}
