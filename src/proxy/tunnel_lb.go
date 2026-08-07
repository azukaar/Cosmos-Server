package proxy

import (
	"sync"
	"sync/atomic"
	"strings"

	"github.com/azukaar/cosmos-server/src/constellation"
	"github.com/azukaar/cosmos-server/src/utils"
)

type routeCounter struct {
	val atomic.Uint64
}

type TunnelLoadBalancer struct {
	counters sync.Map // routeName -> *routeCounter
}

func (lb *TunnelLoadBalancer) getCounter(routeName string) *routeCounter {
	if v, ok := lb.counters.Load(routeName); ok {
		return v.(*routeCounter)
	}
	c := &routeCounter{}
	actual, _ := lb.counters.LoadOrStore(routeName, c)
	return actual.(*routeCounter)
}

// Select picks a key from the given list using the configured LB mode.
// Keys can be numeric indices ("0","1") for regular routes or device names for constellation tunnels.
func (lb *TunnelLoadBalancer) Select(keys []string, routeName string, mode string, sticky bool, stickyKey string) string {
	if len(keys) == 0 {
		return ""
	}

	// Check sticky assignment
	if sticky && stickyKey != "" {
		prev, ok := constellation.GetStickyTarget(stickyKey)
		if ok {
			for _, k := range keys {
				if k == prev {
					return k
				}
			}
			// Sticky target gone from list, fall through to re-assign
		}
	}

	// Mode selection
	var selected string
	switch strings.ToLower(mode) {
	case "round_robin":
		c := lb.getCounter(routeName)
		idx := c.val.Add(1) - 1
		selected = keys[idx%uint64(len(keys))]
	default: // "first" or ""
		selected = keys[0]
	}

	// Store sticky assignment
	if sticky && stickyKey != "" {
		constellation.SetStickyTarget(stickyKey, selected)
	}

	return selected
}

// SelectTarget is a convenience wrapper for constellation tunnel targets.
func (lb *TunnelLoadBalancer) SelectTarget(targets []utils.TunnelTarget, routeName string, mode string, sticky bool, stickyKey string) *utils.TunnelTarget {
	if len(targets) == 0 {
		return nil
	}

	keys := make([]string, len(targets))
	for i, t := range targets {
		keys[i] = t.DeviceName
	}

	key := lb.Select(keys, routeName, mode, sticky, stickyKey)
	if key == "" {
		return nil
	}

	for i := range targets {
		if targets[i].DeviceName == key {
			return &targets[i]
		}
	}
	return nil
}

var DefaultTunnelLB = &TunnelLoadBalancer{}
