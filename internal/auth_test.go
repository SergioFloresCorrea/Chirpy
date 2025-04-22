package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "LumianLee"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}

	// Use bcrypt to compare the hash with the original password
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("Hashed password does not match original password: %v", err)
	}
}

func TestPasswordMismatch(t *testing.T) {
	originalPassword := "LumianLee"
	wrongPassword := "KleinMoretti"

	// Hash the original password
	hashedPassword, err := HashPassword(originalPassword)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}

	// Check that the wrong password does not match
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(wrongPassword))
	if err == nil {
		t.Error("Expected error when comparing wrong password, got nil")
	}
}
