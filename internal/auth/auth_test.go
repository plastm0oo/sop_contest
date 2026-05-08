package auth

import (
	"testing"
	"time"
)

func TestHashPasswordAndCheckPassword(t *testing.T) {
	password := "Secret123"

	hash, err := HashPassword(password, 10)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty password hash")
	}

	if hash == password {
		t.Fatal("password hash must not equal plain password")
	}

	if !CheckPassword(hash, password) {
		t.Fatal("expected password check to succeed")
	}

	if CheckPassword(hash, "Wrong123") {
		t.Fatal("expected password check to fail for wrong password")
	}
}

func TestGenerateAndHashRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty refresh token")
	}

	hash := HashRefreshToken(token)

	if hash == "" {
		t.Fatal("expected non-empty refresh token hash")
	}

	if hash == token {
		t.Fatal("refresh token hash must not equal plain token")
	}

	if len(hash) != 64 {
		t.Fatalf("expected sha256 hex hash length 64, got %d", len(hash))
	}
}

func TestGenerateAndParseAccessToken(t *testing.T) {
	secret := "01234567890123456789012345678901"

	token, err := GenerateAccessToken(
		42,
		"user@hse.ru",
		"admin",
		secret,
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := ParseAccessToken(token, secret)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("expected userID=42, got %d", claims.UserID)
	}

	if claims.Email != "user@hse.ru" {
		t.Fatalf("expected email=user@hse.ru, got %s", claims.Email)
	}

	if claims.Role != "admin" {
		t.Fatalf("expected role=admin, got %s", claims.Role)
	}
}

func TestParseAccessTokenWithWrongSecret(t *testing.T) {
	token, err := GenerateAccessToken(
		42,
		"user@hse.ru",
		"user",
		"01234567890123456789012345678901",
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	_, err = ParseAccessToken(token, "wrong-secret-wrong-secret-wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong JWT secret")
	}
}
