package main

import (
	"testing"
	"time"

	"github.com/MinhBuiAnh/chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := auth.MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned an error: %v", err)
	}
	if token == "" {
		t.Fatal("MakeJWT returned an empty token")
	}

	gotUserID, err := auth.ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned an error for a valid token: %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("ValidateJWT returned user ID %v, want %v", gotUserID, userID)
	}
}

func TestMakeJWTRejectsWrongSecret(t *testing.T) {
	token, err := auth.MakeJWT(uuid.New(), "test-secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned an error: %v", err)
	}

	if _, err := auth.ValidateJWT(token, "wrong-secret"); err == nil {
		t.Fatal("ValidateJWT accepted a token signed with a different secret")
	}
}
