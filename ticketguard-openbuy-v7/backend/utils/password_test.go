package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("StrongPass123!")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "StrongPass123!" {
		t.Fatal("password must not be stored in plain text")
	}
	if !CheckPassword("StrongPass123!", hash) {
		t.Fatal("valid password did not match")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("invalid password matched")
	}
}
