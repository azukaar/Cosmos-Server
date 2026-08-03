package constellation

import (
	"github.com/azukaar/cosmos-server/src/pro"
	"github.com/azukaar/cosmos-server/src/utils"
)

// StartSchedulerInConstellation is the thin bridge that wires pro.StartScheduler
// into the constellation lifecycle. It exists here (not inline at the call site)
// so the `pro` package stays independent of `constellation` — the same pattern
// used by `api_deployments.go` for the HTTP handlers.
//
// This must be called AFTER the NATS client is connected (nc, js are non-nil)
// AND after pro.ClientHeartbeatInit has had a chance to create/attach the
// constellation-deployments KV. `ClientHeartbeatInit` in tunnels.go is the
// right place.
func StartSchedulerInConstellation() {
	device, err := GetCurrentDevice()
	if err != nil {
		utils.Warn("[SCHED] cannot start scheduler: GetCurrentDevice failed: " + err.Error())
		return
	}
	self := sanitizeNATSUsername(device.DeviceName)

	// Registry of placement strategies selectable per-Deployment via
	// Deployment.Strategy. DefaultStrategies() ships round-robin (default)
	// and least-busy; least-busy falls back to round-robin when monitoring
	// is unavailable.
	pro.StartScheduler(&clientConfigLock, js, nc, self, pro.DefaultStrategies())
}

// StopSchedulerInConstellation halts the scheduler. Called from StopHeartbeat
// and CloseNATSClient so the scheduler shuts down cleanly when the
// constellation client disconnects.
func StopSchedulerInConstellation() {
	pro.StopScheduler()
}

// GetCurrentLeaderName returns the sanitized device name of the current
// scheduler leader, or "" when it cannot be determined (no cluster, NATS not
// connected, or no leader elected yet). Best-effort — never errors — so it is
// safe to call from request handlers regardless of cluster state. The name is
// returned exactly as stored (sanitized via sanitizeNATSUsername); callers
// match it client-side rather than reversing the sanitization.
func GetCurrentLeaderName() string {
	name, ok := pro.GetLeaderName(&clientConfigLock, js)
	if !ok {
		return ""
	}
	return name
}
