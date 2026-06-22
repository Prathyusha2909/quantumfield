package auth

import (
	"testing"
	"time"

	"github.com/Prathyusha2909/quantumfield/internal/models"
)

func TestPasswordHashAndVerification(t *testing.T) {
	service := New("test-secret-that-is-long-enough-for-hmac", time.Hour)
	hash, err := service.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("password was not hashed")
	}
	if err := service.VerifyPassword(hash, "correct horse battery staple"); err != nil {
		t.Fatalf("valid password rejected: %v", err)
	}
	if err := service.VerifyPassword(hash, "wrong password"); err == nil {
		t.Fatal("invalid password accepted")
	}
}

func TestTokenRoundTripPreservesIdentityAndRole(t *testing.T) {
	service := New("test-secret-that-is-long-enough-for-hmac", time.Hour)
	token, err := service.CreateToken("user-123", models.RoleAnalyst)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := service.ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-123" || claims.Subject != "user-123" {
		t.Fatalf("unexpected identity claims: %+v", claims)
	}
	if claims.Role != models.RoleAnalyst {
		t.Fatalf("expected role %s, got %s", models.RoleAnalyst, claims.Role)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	service := New("test-secret-that-is-long-enough-for-hmac", -time.Minute)
	token, err := service.CreateToken("user-123", models.RoleAnalyst)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ParseToken(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
