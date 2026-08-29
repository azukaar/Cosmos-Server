package utils

import (
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateSecret returns a cryptographically-random secret encryption key in
// hex form: 32 random bytes encoded as 64 lowercase hex characters. This is
// the format expected by apps such as Homarr (SECRET_ENCRYPTION_KEY) and is
// safe to use as a symmetric encryption key. It is backed by crypto/rand, so
// unlike GenerateRandomString it is not reproducible and is suitable for
// secrets. The returned value never fails (crypto/rand errors are fatal).
func GenerateSecret() string {
	return hex.EncodeToString(GenerateRandomBytes(32))
}

// GenerateSecrets returns n independent, cryptographically-random secret
// encryption keys (see GenerateSecret). Every element is a distinct random
// value, so a template can use {Secrets.0}, {Secrets.1}, ... and each occurrence
// resolves to a unique key.
func GenerateSecrets(n int) []string {
	secrets := make([]string, n)
	for i := range secrets {
		secrets[i] = GenerateSecret()
	}
	return secrets
}

// GenerateRandomBytes returns n cryptographically-random bytes from
// crypto/rand. It panics if the OS entropy source fails, which on any sane
// system only happens for programming errors (e.g. n < 0 or an
// uninitialised crypto stack), so callers can rely on a valid result.
func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := crand.Read(b); err != nil {
		panic(fmt.Sprintf("utils: failed to read crypto/rand: %v", err))
	}
	return b
}
