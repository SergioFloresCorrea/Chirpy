package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SergioFloresCorrea/Chirpy/internal/database"
	"github.com/google/uuid"
)

func ServerReady(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK\n")
}

func (cfg *apiConfig) ValidateAndSaveChirp(w http.ResponseWriter, req *http.Request) {
	type ExpectedJson struct {
		Body string `json:"body"`
	}

	type ResponseJson struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
	expectedJson := ExpectedJson{}
	if err := decoder.Decode(&expectedJson); err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}
	defer req.Body.Close()

	/// check length
	if len(expectedJson.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	validJson := ResponseJson{CleanedBody: replaceBadWords(expectedJson.Body)}
	respondWithJSON(w, http.StatusOK, validJson)
}

func (cfg *apiConfig) CreateUser(w http.ResponseWriter, req *http.Request) {
	type ExpectedJson struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	expectedJson := ExpectedJson{}
	if err := decoder.Decode(&expectedJson); err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}
	defer req.Body.Close()

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     expectedJson.Email,
	}

	cfg.dbQueries.CreateUser(req.Context(), params)
	responseJson := MapCreateUserParamsToUser(params)
	respondWithJSON(w, 201, responseJson)
}
