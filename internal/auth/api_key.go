package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authorizationString := headers.Get("Authorization")

	if len(authorizationString) == 0 {
		return "", fmt.Errorf("Authorization header not found")
	}

	apiKey, found := strings.CutPrefix(authorizationString, "ApiKey ")
	if !found {
		return "", fmt.Errorf("Authorization string is in wrong format")
	}

	return apiKey, nil
}