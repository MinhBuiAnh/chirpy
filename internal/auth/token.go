package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingMethod := jwt.SigningMethodHS256
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	userID, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(userID)
}

func GetBearerToken(headers http.Header) (string, error) {
	authorizationString := headers.Get("Authorization")

	if len(authorizationString) == 0 {
		return "", fmt.Errorf("Authorization header not found")
	}

	tokenString, found := strings.CutPrefix(authorizationString, "Bearer ")
	if !found {
		return "", fmt.Errorf("Authorization string is in wrong format")
	}

	return tokenString, nil
}

func MakeRefreshToken() string {
	data := make([]byte, 32)
	rand.Read(data)
	return hex.EncodeToString(data)
}