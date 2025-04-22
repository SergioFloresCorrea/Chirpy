package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization header present")
	}
	prefix := "ApiKey "
	apiKey := strings.TrimPrefix(authHeader, prefix)
	apiKey = strings.TrimSpace(apiKey)
	return apiKey, nil
}
