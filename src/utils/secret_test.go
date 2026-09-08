package utils

import (
	"regexp"
	"testing"
)

func TestGenerateSecret(t *testing.T) {
	hexRe := regexp.MustCompile(`^[0-9a-f]{64}$`)

	s1 := GenerateSecret()
	if !hexRe.MatchString(s1) {
		t.Fatalf("GenerateSecret() = %q, want 64 lowercase hex chars", s1)
	}
	if len(s1) != 64 {
		t.Fatalf("GenerateSecret() length = %d, want 64", len(s1))
	}

	// Two calls must not collide.
	s2 := GenerateSecret()
	if s1 == s2 {
		t.Fatalf("GenerateSecret() returned the same value twice: %q", s1)
	}
}

func TestGenerateSecrets(t *testing.T) {
	hexRe := regexp.MustCompile(`^[0-9a-f]{64}$`)

	secrets := GenerateSecrets(16)
	if len(secrets) != 16 {
		t.Fatalf("GenerateSecrets(16) length = %d, want 16", len(secrets))
	}

	seen := map[string]bool{}
	for _, s := range secrets {
		if !hexRe.MatchString(s) {
			t.Fatalf("GenerateSecrets() produced non-hex value: %q", s)
		}
		if seen[s] {
			t.Fatalf("GenerateSecrets() produced duplicate value: %q", s)
		}
		seen[s] = true
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	b := GenerateRandomBytes(32)
	if len(b) != 32 {
		t.Fatalf("GenerateRandomBytes(32) length = %d, want 32", len(b))
	}
	// Non-empty (random bytes could theoretically be all zeros, but with 32
	// bytes the chance is ~2^-256; treat as a sanity check that reads work).
	sum := 0
	for _, x := range b {
		sum += int(x)
	}
	if sum == 0 {
		t.Fatalf("GenerateRandomBytes(32) returned all zeros")
	}
}
