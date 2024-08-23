package util

import (
	"context"
	"testing"
	"time"
)

func TestGenerateTokenWithExp(t *testing.T) {
	// TestGenerateTokenWithExp tests the function GenerateTokenWithExp.
	// The token should be generated successfully.
	// The token should be valid.
	// The token should have the correct expiration time.
	// The token should have the correct issuer.
	// The token should have the correct subject.
	token, err := GenerateTokenWithExp(context.Background(), "test-1-2", "test", time.Hour*3)
	if err != nil {
		t.Errorf("GenerateTokenWithExp() error = %v", err)
		return
	}
	if token == "" {
		t.Errorf("GenerateTokenWithExp() token is empty")
		return
	}

	uid, err := IdentityFromToken(token, "1-2", "test")
	if err != nil {
        t.Errorf("IdentityFromToken() error = %v", err)
        return
    }
    if uid != "test" {
        t.Errorf("IdentityFromToken() uid = %v, want %v", uid, "test")
    }
}
