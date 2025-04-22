package auth

import (
	"log"
	"unsafe"

	"golang.org/x/crypto/bcrypt"
)

func StringToBytes(string_ string) (bytes []byte) {
	return unsafe.Slice(unsafe.StringData(string_), len(string_))
}

func BytesToString(bytes []byte) string {
	return unsafe.String(unsafe.SliceData(bytes), len(bytes))
}

func HashPassword(password string) (string, error) {
	passwordBytes := StringToBytes(password)
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		log.Printf("An error ocurred during hashing process: %v", err)
		return "", err
	}
	hashedPassword := BytesToString(hashedPasswordBytes)
	return hashedPassword, nil
}

func CheckPasswordHash(hash, password string) error {
	hashBytes := StringToBytes(hash)
	passwordBytes := StringToBytes(password)
	if err := bcrypt.CompareHashAndPassword(hashBytes, passwordBytes); err != nil {
		log.Printf("The password is incorrect: %v", err)
		return err
	}
	return nil
}
