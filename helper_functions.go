package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	auth "github.com/SergioFloresCorrea/Chirpy/internal"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func replaceBadWords(body string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	cleanedBody := make([]string, 0, len(body))
	words := strings.Split(body, " ")
	for _, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			cleanedBody = append(cleanedBody, "****")
		} else {
			cleanedBody = append(cleanedBody, word)
		}
	}
	return strings.Join(cleanedBody, " ")
}

func hasNoBody(r *http.Request) bool {
	// Check if Body is nil or ContentLength is zero or less
	return r.Body == nil || r.ContentLength <= 0
}

func checkAuthHeader(r *http.Request) (string, error) {
	headers := r.Header
	tokenString, err := auth.GetBearerToken(headers)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
