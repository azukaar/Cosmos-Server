package docker

import (
	"strings"

	"github.com/azukaar/cosmos-server/src/utils"
)

// NetworkModeContainerRef returns true when the given docker network mode is a
// "container:<name-or-id>" reference, i.e. this container shares the network
// namespace of another container.
func NetworkModeContainerRef(mode string) bool {
	return strings.HasPrefix(mode, "container:")
}

// NetworkModeServiceRef returns true when the given docker network mode is a
// "service:<name>" reference (compose-style alias for a container reference).
func NetworkModeServiceRef(mode string) bool {
	return strings.HasPrefix(mode, "service:")
}

// NetworkModeRefTarget extracts the referenced container name-or-ID from a
// "container:<...>" or "service:<...>" network mode. Returns "" when the mode
// is not a container/service reference.
func NetworkModeRefTarget(mode string) string {
	if NetworkModeContainerRef(mode) {
		return strings.TrimPrefix(mode, "container:")
	}
	if NetworkModeServiceRef(mode) {
		return strings.TrimPrefix(mode, "service:")
	}
	return ""
}

// ContainerRefToName resolves a "container:<ref>" / "service:<ref>" network
// mode target to its stable container name (IDs go stale on recreate; names
// do not). Unresolvable refs are returned unchanged.
func ContainerRefToName(mode string) string {
	target := NetworkModeRefTarget(mode)
	if target == "" {
		return mode
	}

	name := ResolveContainerRefToName(target)
	if name == "" {
		// Unresolvable right now: keep the original value untouched.
		return mode
	}

	// always persist container:<name> (service: and IDs go stale on recreate)
	return "container:" + name
}

// ResolveContainerRefToName resolves a container name/ID/prefix to its stable
// name, or "" when not found / Docker unreachable.
func ResolveContainerRefToName(ref string) string {
	if ref == "" {
		return ""
	}

	if err := Connect(); err != nil {
		utils.Debug("ResolveContainerRefToName: docker connect failed: " + err.Error())
		return ""
	}

	container, err := DockerClient.ContainerInspect(DockerContext, ref)
	if err != nil {
		utils.Debug("ResolveContainerRefToName: cannot inspect " + ref + ": " + err.Error())
		return ""
	}

	return strings.TrimPrefix(container.Name, "/")
}
