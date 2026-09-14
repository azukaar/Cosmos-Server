package utils

// This file covers the clean-install "wipe" (v0.23.0 SQLite era).
//
// auth.db is opened once at process start and held open, so any clean-install
// step that empties the whole config folder (the old container-mode
// `os.RemoveAll("/config")`) orphans the store: every write made afterwards -
// including the admin user created a few steps later - lands in unlinked
// inodes and vanishes on the next restart.
//
// The fix removes only the config file (like service installs always did),
// leaving the SQLite stores and derived state alone. The store handle stays
// valid, so the setup flow's own DeleteAllUsersLocal + CreateUser persist.

import (
	"os"
	"testing"
	"time"
)

// The fixed container clean-install flow:
//  1. InitStore() opened <CONFIGFOLDER>/auth.db at startup
//  2. clean-install removes ONLY the config file - the folder (and the open
//     store) is left untouched, exactly like a service install
//  3. wizard wipes users and creates the admin through the same open handle
//  4. a later "restart" (re-InitStore) must still see the admin
func TestContainerCleanInstallWithFix(t *testing.T) {
	dir := t.TempDir()
	CONFIGFOLDER = dir + string(os.PathSeparator)
	defer func() { CONFIGFOLDER = "/var/lib/cosmos/" }()

	if err := InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	// step "-1": only cosmos.config.json is removed (same in container and service)
	configFile := CONFIGFOLDER + "cosmos.config.json"
	if err := os.WriteFile(configFile, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(configFile); err != nil {
		t.Fatal(err)
	}

	// sanity: the folder (and therefore the open store) is still there
	if _, err := os.Stat(CONFIGFOLDER); err != nil {
		t.Fatalf("config folder must not be removed: %v", err)
	}

	// step 4: local wipe + admin creation
	if err := DeleteAllUsersLocal(); err != nil {
		t.Fatalf("DeleteAllUsersLocal: %v", err)
	}
	if err := CreateUser(User{
		Nickname: "admin", Password: "x", Role: ADMIN,
		PasswordCycle: 0, CreatedAt: time.Now(), RegisteredAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// restart
	if err := InitStore(); err != nil {
		t.Fatalf("restart InitStore: %v", err)
	}
	if _, err := GetUser("admin"); err != nil {
		t.Fatalf("admin user LOST after clean-install + restart: %v", err)
	}
	t.Log("PASS: admin user persists after config-only clean install + restart")
}

// Documents the old failure: emptying the config folder while the store is
// open orphans the admin write. Not a fix - a regression guard explaining why
// the narrow wipe is important.
func TestContainerCleanInstallBugBaseline(t *testing.T) {
	dir := t.TempDir()
	CONFIGFOLDER = dir + string(os.PathSeparator)
	defer func() { CONFIGFOLDER = "/var/lib/cosmos/" }()

	if err := InitStore(); err != nil {
		t.Fatalf("InitStore: %v", err)
	}

	// OLD container behavior: RemoveAll(CONFIGFOLDER) while store is open
	os.RemoveAll(dir)
	os.Mkdir(dir, 0700)

	if err := CreateUser(User{
		Nickname: "admin", Password: "x", Role: ADMIN,
		PasswordCycle: 0, CreatedAt: time.Now(), RegisteredAt: time.Now(),
	}); err != nil {
		t.Logf("CreateUser errored (db gone) - also a symptom")
		return
	}
	if err := InitStore(); err != nil {
		t.Logf("restart InitStore fails: %v", err)
		return
	}
	if _, err := GetUser("admin"); err == nil {
		t.Logf("NOTE: user survived in this run (inode reuse)")
	} else {
		t.Logf("BASELINE: admin user lost with the old RemoveAll wipe (GetUser: %v)", err)
	}
}
