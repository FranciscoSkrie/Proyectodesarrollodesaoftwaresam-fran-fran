package domain

import "testing"

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  USER@Example.COM "); got != "user@example.com" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeRole(t *testing.T) {
	if NormalizeRole(RoleAdmin) != RoleAdmin {
		t.Fatal("admin role was not preserved")
	}
	if NormalizeRole("unexpected") != RoleUsuario {
		t.Fatal("unknown role must normalize to usuario")
	}
}

func TestIsValidRole(t *testing.T) {
	if !IsValidRole(RoleUsuario) || !IsValidRole(RoleAdmin) || IsValidRole("root") {
		t.Fatal("unexpected role validation result")
	}
}

func TestOfferValidate(t *testing.T) {
	valid := Offer{EventID: 1, Title: "Oferta", Price: 100, Quantity: 1}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid offer rejected: %v", err)
	}
	invalid := Offer{EventID: 1, Title: "Oferta", Price: 0, Quantity: 1}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected invalid price error")
	}
}
