package constellation

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/azukaar/cosmos-server/src/utils"
)

// repoRootDir returns the absolute repository root (two levels up from this package)
func repoRootDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("testenv: cannot locate source file")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
}

// setupTestEnv points CONFIGFOLDER at a temp dir, loads a fresh config and resets
// all constellation package globals. mutate (optional) tweaks the config before load.
// Tests using this must NOT run in parallel: they share package-level state.
func setupTestEnv(t *testing.T, mutate func(*utils.Config)) string {
	t.Helper()

	tmp := t.TempDir() + "/"
	prevFolder := utils.CONFIGFOLDER
	utils.CONFIGFOLDER = tmp

	cfg := utils.DefaultConfig
	cfg.ConstellationConfig.Enabled = true
	cfg.ConstellationConfig.DNSFallback = "8.8.8.8:53"
	if mutate != nil {
		mutate(&cfg)
	}
	utils.LoadBaseMainConfig(cfg)

	utils.CloseEmbeddedDB()
	resetConstellationGlobals()

	t.Cleanup(func() {
		utils.CloseEmbeddedDB()
		resetConstellationGlobals()
		utils.CONFIGFOLDER = prevFolder
	})

	return tmp
}

func resetConstellationGlobals() {
	cachedCurrentDevice = nil
	CachedDevices = map[string]utils.ConstellationDevice{}
	CachedDeviceNames = map[string]string{}
	NebulaStarted = false
	NebulaHasStarted = false
	NATSStarted = false

	dnsMux.Lock()
	DNSStarted = false
	dnsServer = nil
	dnsStarting = false
	dnsMux.Unlock()
	DNSBlacklist = map[string]bool{}
}

// writeNebulaYML writes CONFIGFOLDER/nebula.yml from the given fields and
// invalidates the current-device cache. Typical fields: cstln_device_name,
// cstln_api_key, cstln_ip, cstln_is_lighthouse...
func writeNebulaYML(t *testing.T, fields map[string]interface{}) {
	t.Helper()

	data, err := yaml.Marshal(fields)
	if err != nil {
		t.Fatal("testenv: marshal nebula.yml:", err)
	}
	if err := os.WriteFile(utils.CONFIGFOLDER+"nebula.yml", data, 0600); err != nil {
		t.Fatal("testenv: write nebula.yml:", err)
	}
	cachedCurrentDevice = nil
}

// chdirWithCertBinary moves CWD to a temp dir containing symlinks to the
// nebula-cert/nebula binaries, since the code resolves them as "./nebula-cert".
// Cert files generated with saveToFile=false also land in this CWD.
func chdirWithCertBinary(t *testing.T) string {
	t.Helper()

	root := repoRootDir(t)
	work := t.TempDir()

	for _, bin := range []string{"nebula-cert", "nebula", "nebula-arm-cert", "nebula-arm"} {
		src := filepath.Join(root, bin)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := os.Symlink(src, filepath.Join(work, bin)); err != nil {
			t.Fatal("testenv: symlink "+bin+":", err)
		}
	}

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal("testenv: getwd:", err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatal("testenv: chdir:", err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	return work
}
