package proxy

import (
	"sync"
	"sync/atomic"

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

func (lb *TunnelLoadBalancer) SelectTarget(targets []utils.TunnelTarget, routeName string, mode string, sticky bool, stickyKey string) *utils.TunnelTarget {
	if len(targets) == 0 {
		return nil
	}

	// Check sticky assignment
	if sticky && stickyKey != "" {
		deviceName, ok := constellation.GetStickyTarget(stickyKey)
		if ok {
			for i := range targets {
				if targets[i].DeviceName == deviceName {
					return &targets[i]
				}
			}
			// Sticky device gone from targets, fall through to re-assign
		}
	}

	// Mode selection
	var selected *utils.TunnelTarget
	switch mode {
	case "round_robin":
		c := lb.getCounter(routeName)
		idx := c.val.Add(1) - 1
		selected = &targets[idx%uint64(len(targets))]
	default: // "first" or ""
		selected = &targets[0]
	}

	// Store sticky assignment
	if sticky && stickyKey != "" && selected != nil {
		constellation.SetStickyTarget(stickyKey, selected.DeviceName)
	}

	return selected
}

var DefaultTunnelLB = &TunnelLoadBalancer{}
