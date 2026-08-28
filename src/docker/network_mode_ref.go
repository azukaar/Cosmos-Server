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

// ContainerRefToName resolves the target of a "container:<...>" /
// "service:<...>" network mode to the target container's stable name (without
// the leading "/"). Non-reference modes are returned unchanged.
//
// This is the heart of the container-mode durability fix. Docker accepts both
// a name and an ID in "container:<ref>", but the daemon does NOT rewrite the
// stored reference when the target container is recreated — the old ID then
// points at a deleted container and the network sharing silently breaks
// (commonly after an update/recreate of the referenced container). The
// container name, on the other hand, is deliberately stable across recreations
// (unless the user actively renames it), so we always persist/resolve to the
// name.
//
// If the target cannot be resolved (Docker unreachable, or the container does
// not exist yet — e.g. a compose stack mid-creation), the input is returned
// unchanged so a valid setup is never broken by a failed normalization.
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

	// Always normalize to container:<name> — never service:<...>, never an ID.
	// service: is a compose-only convenience that Docker itself turns into
	// container:<id> at create time; persisting a container name makes the
	// reference survive recreations of both containers.
	return "container:" + name
}

// ResolveContainerRefToName resolves a container name / full ID / unambiguous
// ID prefix to its stable name (without the leading "/"). Returns "" when the
// container cannot be found or Docker is unreachable — callers decide whether
// to fall back to the original input.
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
