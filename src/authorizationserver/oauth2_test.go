package authorizationserver

import (
	"testing"

	"github.com/azukaar/cosmos-server/src/utils"
)

func TestLoadSigningKey(t *testing.T) {
	previous := utils.CONFIGFOLDER
	utils.CONFIGFOLDER = t.TempDir() + "/"
	t.Cleanup(func() { utils.CONFIGFOLDER = previous })

	first := loadSigningKey()
	second := loadSigningKey()
	if !first.Equal(second) {
		t.Fatal("the signing key changed between two loads")
	}
	if id := keyID(first); id == "" || id != keyID(second) {
		t.Fatalf("unstable key id: %q", id)
	}

	config := utils.ReadConfigFromFile()
	config.HTTPConfig.OIDCSigningKey = "not a key"
	config.HTTPConfig.OIDCSigningKeyShared = true
	utils.SetBaseMainConfig(config)

	replaced := loadSigningKey()
	if replaced.Equal(first) {
		t.Fatal("an unreadable key was not replaced")
	}
	if keyID(replaced) == keyID(first) {
		t.Fatal("a new key must get a new id")
	}
	if utils.ReadConfigFromFile().HTTPConfig.OIDCSigningKeyShared {
		t.Fatal("a new key is not shared with the cluster yet")
	}
}

// every generated key starts with the same characters
func TestDerivePKCEVerifierUsesTheSecretPartOfTheKey(t *testing.T) {
	previous := utils.GetMainConfig()
	t.Cleanup(func() { utils.LoadBaseMainConfig(previous) })

	verifier := func(key string) string {
		config := previous
		config.HTTPConfig.AuthPrivateKey = "-----BEGIN PRIVATE KEY-----\nMC4C" + key
		utils.LoadBaseMainConfig(config)
		return utils.DerivePKCEVerifier("app.example.com", "/")
	}

	if verifier("AAAA") == verifier("BBBB") {
		t.Fatal("two different keys gave the same verifier")
	}
}

func TestGenerateRandomString(t *testing.T) {
	if got := utils.GenerateRandomString(32); len(got) != 32 || got == utils.GenerateRandomString(32) {
		t.Fatalf("got %q", got)
	}
}
