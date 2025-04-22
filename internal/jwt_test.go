package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "LumianLee"
	expiresIn := time.Duration(5 * time.Second)
	ss, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() returned error: %v", err)
	}

	DecoderUserID, err := ValidateJWT(ss, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT() returned error: %v", err)
	}

	if DecoderUserID != userID {
		t.Errorf("Decoded user ID %v does not match the input user id %v", DecoderUserID, userID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "LumianLee"
	expiresIn := time.Duration(-1 * time.Second) // already expired

	ss, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() returned error: %v", err)
	}

	_, err = ValidateJWT(ss, tokenSecret)
	if err == nil {
		t.Fatal("ValidateJWT() should return error for expired token, but got none")
	}
}

func TestValidateJWT_InvalidSecret(t *testing.T) {
	userID := uuid.New()
	originalSecret := "LumianLee"
	wrongSecret := "NotTheRightOne"
	expiresIn := 5 * time.Second

	ss, err := MakeJWT(userID, originalSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() returned error: %v", err)
	}

	_, err = ValidateJWT(ss, wrongSecret)
	if err == nil {
		t.Fatal("ValidateJWT() should return error for wrong secret, but got none")
	}
}

func TestBearerToken(t *testing.T) {
	req, err := http.NewRequest("GET", "https://api.example.com/data", nil)
	if err != nil {
		t.Fatalf("Error creating a request")
	}

	ExpectedAuthToken := "token123"
	req.Header.Set("Authorization", "Bearer "+ExpectedAuthToken)

	authToken, err := GetBearerToken(req.Header)
	if err != nil {
		t.Fatalf("GetBearerToken() returnen error: %v", err)
	}

	if authToken != ExpectedAuthToken {
		t.Errorf("Mismatch in expected %s and received %s authentication tokens.", ExpectedAuthToken, authToken)
	}
}
