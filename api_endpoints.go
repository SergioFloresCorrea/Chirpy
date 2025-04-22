package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	auth "github.com/SergioFloresCorrea/Chirpy/internal"
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

	tokenString, err := checkAuthHeader(req)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
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

	params := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      replaceBadWords(expectedJson.Body),
		UserID:    userID,
	}

	chirp, err := cfg.dbQueries.CreateChirp(req.Context(), params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}
	responseJson := Chirp(chirp)
	respondWithJSON(w, 201, responseJson)
}

func (cfg *apiConfig) CreateUser(w http.ResponseWriter, req *http.Request) {
	type ExpectedJson struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type ResponseJson struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	expectedJson := ExpectedJson{}
	if err := decoder.Decode(&expectedJson); err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}
	defer req.Body.Close()

	hashedPassword, err := auth.HashPassword(expectedJson.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error in hashing password")
		return
	}
	params := database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          expectedJson.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.dbQueries.CreateUser(req.Context(), params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}
	responseJson := ResponseJson{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, responseJson)
}

func (cfg *apiConfig) LoginUser(w http.ResponseWriter, req *http.Request) {
	type ExpectedJson struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type ResponseJson struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	decoder := json.NewDecoder(req.Body)
	expectedJson := ExpectedJson{}
	if err := decoder.Decode(&expectedJson); err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}
	defer req.Body.Close()

	user, err := cfg.dbQueries.GetUserByEmail(req.Context(), expectedJson.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := auth.CheckPasswordHash(user.HashedPassword, expectedJson.Password); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenString, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error ocurred in creating a JWT")
		return
	}

	refreshToken, err := auth.MakeRefreshToken() // the error is always nil
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}

	params := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}

	_, err = cfg.dbQueries.CreateRefreshToken(req.Context(), params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}

	responseJson := ResponseJson{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, responseJson)
}

func (cfg *apiConfig) RefreshAccessToken(w http.ResponseWriter, req *http.Request) {
	type ResponseJson struct {
		Token string `json:"token"`
	}

	if !hasNoBody(req) {
		respondWithError(w, 400, "request body not allowed")
		return
	}

	tokenRefreshString, err := checkAuthHeader(req)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}

	tokenRefreshDb, err := cfg.dbQueries.GetRefreshTokenByToken(req.Context(), tokenRefreshString)
	if err != nil || tokenRefreshDb.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if tokenRefreshDb.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(req.Context(), tokenRefreshString)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}

	tokenString, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "An error ocurred in creating a JWT")
		return
	}

	respondWithJSON(w, http.StatusOK, ResponseJson{Token: tokenString})
}

func (cfg *apiConfig) RevokeRefreshToken(w http.ResponseWriter, req *http.Request) {
	if !hasNoBody(req) {
		respondWithError(w, 400, "request body not allowed")
		return
	}

	tokenRefreshString, err := checkAuthHeader(req)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}

	params := database.SetRevokeAtParams{
		RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
		Token:     tokenRefreshString,
	}
	err = cfg.dbQueries.SetRevokeAt(req.Context(), params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}

	respondWithJSON(w, 204, nil)
}

func (cfg *apiConfig) GetAllChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(req.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}
	responseJson := make([]Chirp, 0, len(chirps))
	for _, chirp := range chirps {
		responseJson = append(responseJson, Chirp(chirp))
	}
	respondWithJSON(w, 200, responseJson)
}

func (cfg *apiConfig) GetChirpByID(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("%v", err))
		return
	}
	responseJson := Chirp(chirp)
	respondWithJSON(w, http.StatusOK, responseJson)
}

func (cfg *apiConfig) DeleteChirpByID(w http.ResponseWriter, req *http.Request) {
	accessToken, err := checkAuthHeader(req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("%v", err))
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Unauthorized")
		return
	}

	err = cfg.dbQueries.DeleteChirpByID(req.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong while deleting the chirp")
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) UpdateOwnEmail(w http.ResponseWriter, req *http.Request) {
	type ExpectedJson struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type ResponseJson struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	accessToken, err := checkAuthHeader(req)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(req.Body)
	expectedJson := ExpectedJson{}
	if err := decoder.Decode(&expectedJson); err != nil {
		respondWithError(w, 400, "Something went wrong")
		return
	}
	defer req.Body.Close()

	hashedPassword, err := auth.HashPassword(expectedJson.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}

	params := database.UpdateUserEmailAndPasswordParams{
		Email:          expectedJson.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}
	user, err := cfg.dbQueries.UpdateUserEmailAndPassword(req.Context(), params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("%v", err))
		return
	}

	responseJson := ResponseJson{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, http.StatusOK, responseJson)
}
