package docker

import (
	"encoding/json"

	"github.com/docker/docker/api/types/mount"
)

// CosmosMount mirrors mount.Mount but uses lowercase JSON tags to match
// the Docker Compose specification and the rest of Cosmos's API conventions.
// It also supports reading legacy uppercase fields for backward compatibility.
type CosmosMount struct {
	Type   string `json:"type"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// UnmarshalJSON implements a compatibility layer that accepts both lowercase
// (new canonical) and uppercase (legacy Docker SDK) field names.
func (c *CosmosMount) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*c = CosmosMount{}

	// Helper: try lowercase key first, then uppercase fallback
	getStr := func(low, up string) string {
		if v, ok := raw[low]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		if v, ok := raw[up]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	c.Type = getStr("type", "Type")
	c.Source = getStr("source", "Source")
	c.Target = getStr("target", "Target")
	return nil
}

// ToDockerMount converts a CosmosMount into the Docker SDK mount.Mount type.
func (c CosmosMount) ToDockerMount() mount.Mount {
	return mount.Mount{
		Type:   mount.Type(c.Type),
		Source: c.Source,
		Target: c.Target,
	}
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
	return CosmosMount{
		Type:   string(m.Type),
		Source: m.Source,
		Target: m.Target,
	}
}

// FromDockerMountSlice converts a slice of mount.Mount into a slice of CosmosMount.
func FromDockerMountSlice(mounts []mount.Mount) []CosmosMount {
	c := make([]CosmosMount, len(mounts))
	for i, m := range mounts {
		c[i] = FromDockerMount(m)
	}
	return c
}
