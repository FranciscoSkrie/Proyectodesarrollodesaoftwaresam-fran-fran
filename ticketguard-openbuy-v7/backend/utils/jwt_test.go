package utils

import (
	"testing"
	"ticketguard/backend/domain"
	"time"
)

func TestJWTGenerateAndParse(t *testing.T) {
	token, err := GenerateJWT(42, domain.RoleAdmin, "test-secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseJWT(token, "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.Role != domain.RoleAdmin {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	token, err := GenerateJWT(7, domain.RoleUsuario, "secret-one", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseJWT(token, "secret-two"); err == nil {
		t.Fatal("expected invalid token error")
	}
}
