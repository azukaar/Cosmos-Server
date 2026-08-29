package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func writeZip(t *testing.T, dir string, names ...string) string {
	t.Helper()
	p := filepath.Join(dir, "u.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for _, n := range names {
		e, err := w.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		e.Write([]byte("x"))
	}
	w.Close()
	f.Close()
	return p
}

func TestUnzipRejectsTraversal(t *testing.T) {
	for _, bad := range []string{"cosmos/../../evil", "../evil", "/etc/evil"} {
		tmp := t.TempDir()
		dest := filepath.Join(tmp, "dest")
		os.Mkdir(dest, 0o755)
		z := writeZip(t, tmp, "cosmos/ok.txt", bad)
		if err := unzip(z, dest); err == nil {
			t.Errorf("entry %q accepted, want rejection", bad)
		}
		if _, err := os.Stat(filepath.Join(tmp, "evil")); err == nil {
			t.Errorf("entry %q escaped to parent dir", bad)
		}
	}
}

func TestUnzipHappyPath(t *testing.T) {
	tmp := t.TempDir()
	dest := filepath.Join(tmp, "dest")
	os.Mkdir(dest, 0o755)
	z := writeZip(t, tmp, "cosmos/a.txt", "cosmos/sub/b.txt", "cosmos/cosmos-launcher")
	if err := unzip(z, dest); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"a.txt", "sub/b.txt", "cosmos-launcher.updated"} {
		if _, err := os.Stat(filepath.Join(dest, want)); err != nil {
			t.Errorf("missing %s", want)
		}
	}
}
