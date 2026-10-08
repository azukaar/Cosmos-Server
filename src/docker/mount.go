package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/mount"
)

// CosmosMount mirrors mount.Mount but uses lowercase JSON tags to match
// the Docker Compose specification and the rest of Cosmos's API conventions.
// It also supports reading legacy uppercase fields for backward compatibility.
type CosmosMount struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Target string `json:"target"`
	// SubPath mounts a subdirectory (or a single file) inside a named
	// volume instead of the whole volume, like docker-compose's `subpath`.
	// It is wired to mount.VolumeOptions.Subpath for volume mounts.
	SubPath string `json:"subpath,omitempty"`
	// ReadOnly mounts the source read-only (compose short-syntax `:ro`).
	ReadOnly bool `json:"readOnly,omitempty"`
	// NoCopy is the volume-nocopy mount option (VolumeOptions.NoCopy). It is
	// auto-enabled for subpath mounts on engines older than Docker 29.5.0 as a
	// workaround for moby#52546, but users can also set it explicitly; it is
	// preserved through the export round-trip.
	NoCopy bool `json:"noCopy,omitempty"`
}

// UnmarshalJSON implements a compatibility layer that accepts both lowercase
// (new canonical) and uppercase (legacy Docker SDK) field names, and matches
// all keys case-insensitively so a readOnly/subpath is never silently dropped
// due to casing (e.g. subPath vs subpath, readOnly vs read_only).
func (c *CosmosMount) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*c = CosmosMount{}

	find := func(name string) (interface{}, bool) {
		for k, v := range raw {
			if strings.EqualFold(k, name) {
				return v, true
			}
		}
		return nil, false
	}
	getStr := func(name string) string {
		if v, ok := find(name); ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}
	getBool := func(names ...string) bool {
		for _, n := range names {
			if v, ok := find(n); ok {
				if b, ok := v.(bool); ok {
					return b
				}
			}
		}
		return false
	}

	c.Type = getStr("type")
	c.Source = getStr("source")
	c.Target = getStr("target")
	c.SubPath = getStr("subpath")
	c.ReadOnly = getBool("readOnly", "read_only", "readonly")
	c.NoCopy = getBool("noCopy", "nocopy", "no_copy")
	if v, ok := raw["mode"]; ok {
		if mode, ok := v.(string); ok {
			for _, seg := range strings.Split(mode, ",") {
				if seg == "ro" {
					c.ReadOnly = true
				}
			}
		}
	}
	return nil
}

// ToDockerMount converts a CosmosMount into the Docker SDK mount.Mount type.
func (c CosmosMount) ToDockerMount() mount.Mount {
	m := mount.Mount{
		Type:     mount.Type(c.Type),
		Source:   c.Source,
		Target:   c.Target,
		ReadOnly: c.ReadOnly,
	}

	// Docker engine 26.0+ supports mounting a subpath of a named volume
	// (volume-subpath). docker-compose exposes this as the `subpath` option.
	if c.SubPath != "" {
		// NoCopy (volume-nocopy) is set as a workaround on engines older than
		// Docker 29.5.0: moby#52546 — Docker's seed-from-image copy step treats
		// a file subpath as a directory and fails ("open .../safe-mount...:
		// not a directory"). moby PR #52584 (Docker >= 29.5.0) fixed this
		// daemon-side by skipping population for non-directory subpaths, so on
		// 29.5.0+ NoCopy is only applied when the user explicitly requested it.
		noCopy := c.NoCopy || (!SupportsBuiltinFileSubpathFix())
		m.VolumeOptions = &mount.VolumeOptions{
			Subpath: c.SubPath,
			NoCopy:  noCopy,
		}
	}

	return m
}

// ToDockerMountSlice converts a slice of CosmosMount into a slice of mount.Mount.
func ToDockerMountSlice(c []CosmosMount) []mount.Mount {
	m := make([]mount.Mount, len(c))
	for i, mount := range c {
		m[i] = mount.ToDockerMount()
	}
	return m
}

// IsBindMount returns true if the CosmosMount type is a bind mount.
func (c CosmosMount) IsBindMount() bool {
	return c.Type == string(mount.TypeBind)
}

// IsVolumeMount returns true if the CosmosMount type is a volume mount.
func (c CosmosMount) IsVolumeMount() bool {
	return c.Type == string(mount.TypeVolume)
}

// FromDockerMount converts a Docker SDK mount.Mount into a CosmosMount.
func FromDockerMount(m mount.Mount) CosmosMount {
	cm := CosmosMount{
		Type:     string(m.Type),
		Source:   m.Source,
		Target:   m.Target,
		ReadOnly: m.ReadOnly,
	}
	if m.VolumeOptions != nil {
		cm.SubPath = m.VolumeOptions.Subpath
		cm.NoCopy = m.VolumeOptions.NoCopy
	}
	return cm
}

// FromDockerMountSlice converts a slice of mount.Mount into a slice of CosmosMount.
func FromDockerMountSlice(mounts []mount.Mount) []CosmosMount {
	c := make([]CosmosMount, len(mounts))
	for i, m := range mounts {
		c[i] = FromDockerMount(m)
	}
	return c
}

// EnsureSubpathExists creates the missing parent directories and, when the
// final subpath component looks like a file (basename contains a dot), the
// file itself, inside the given mount root (a named volume's _data dir).
//
// Docker's volume-subpath mount requires the subpath to already exist inside
// the volume; this mirrors what Cosmos already does for bind mounts.
//
// Returns the paths that were created.
func EnsureSubpathExists(mountRoot, subpath string) ([]string, error) {
	created := []string{}
	if subpath == "" {
		return created, nil
	}
	clean := strings.TrimPrefix(filepath.Clean(subpath), string(filepath.Separator))
	if clean == "." || clean == "" {
		return created, nil
	}

	dirs := filepath.Dir(clean)
	if dirs != "." && dirs != "" {
		p := filepath.Join(mountRoot, dirs)
		if err := os.MkdirAll(p, 0o750); err != nil {
			return created, err
		}
		created = append(created, p)
	}

	full := filepath.Join(mountRoot, clean)
	if _, err := os.Stat(full); err == nil {
		return created, nil // already exists (dir or file)
	} else if !os.IsNotExist(err) {
		return created, err
	}

	if looksLikeFile(clean) {
		dir := filepath.Dir(full)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return created, err
		}
		f, err := os.OpenFile(full, os.O_CREATE|os.O_WRONLY, 0o640)
		if err != nil {
			return created, err
		}
		f.Close()
		created = append(created, full)
	} else {
		if err := os.MkdirAll(full, 0o750); err != nil {
			return created, err
		}
		created = append(created, full)
	}

	return created, nil
}

// looksLikeFile reports whether a subpath's basename looks like a file
// (contains a dot), matching how bind-mount auto-creation detects files.
func looksLikeFile(subpath string) bool {
	base := filepath.Base(subpath)
	return strings.Contains(base, ".")
}

// ValidateMountConflicts checks a service's volume list for the ONE mount
// combination Docker's volume-subpath cannot handle: a VOLUME mount whose
// subpath resolves to a FILE, with its target nested inside another mount's
// target directory.
//
// Everything else is intentionally allowed:
//   - plain nested mounts (e.g. -v /a:/a -v /a/b:/a/b) work in Docker
//   - volume + directory subpath works, even nested
//   - volume + file subpath works when NOT nested
//   - bind mounts are never affected (no subpath mechanism)
//
// The failing case is specific: runc's bind of the volume-subpath "safe-mount"
// onto a target that sits inside a directory which is itself another mount's
// target aborts with a cryptic ENOTDIR ("mount a directory onto a file (or
// vice-versa)"). This helper rejects exactly that combination up front.
func ValidateMountConflicts(mounts []CosmosMount) error {
	for i := range mounts {
		m := mounts[i]
		if m.Type != "volume" || m.SubPath == "" || !looksLikeFile(m.SubPath) {
			continue
		}
		targetI := filepath.Clean(m.Target)
		if targetI == "" || targetI == "." || targetI == "/" {
			continue
		}
		for j := range mounts {
			if i == j {
				continue
			}
			targetJ := filepath.Clean(mounts[j].Target)
			if targetJ == "" || targetJ == "." || targetJ == "/" {
				continue
			}
			if isStrictPathInside(targetI, targetJ) {
				return fmt.Errorf(
					"Invalid volume configuration: volume subpath mount target '%s' (source '%s', subpath '%s') is inside another mount's target '%s' (source '%s'). "+
						"Docker cannot mount a VOLUME FILE subpath into a directory that is already another mount point; this fails at container start. "+
						"Move it to a distinct, non-nested target (or mount the whole directory subpath instead of a single file).",
					m.Target, m.Source, m.SubPath,
					mounts[j].Target, mounts[j].Source)
			}
		}
	}
	return nil
}

func isStrictPathInside(child, parent string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
