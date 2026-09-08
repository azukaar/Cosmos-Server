package constellation

import (
	"strconv"

	"github.com/azukaar/cosmos-server/src/pro"
	"github.com/azukaar/cosmos-server/src/utils"
)

func StartSchedulerInConstellation() {
	device, err := GetCurrentDevice()
	if err != nil {
		utils.Warn("[SCHED] cannot start scheduler: GetCurrentDevice failed: " + err.Error())
		return
	}
	self := device.DeviceName

	followerOnly := device.CosmosNode != 2
	if followerOnly {
		utils.Log("[SCHED] node is not a manager (CosmosNode=" + strconv.Itoa(device.CosmosNode) + ") — scheduler starts leader-ineligible")
	}

	// Registry of placement strategies selectable per-Deployment via
	// Deployment.Strategy
	pro.StartSchedulerWithOptions(&clientConfigLock, js, nc, self, pro.DefaultStrategies(), pro.SchedulerOptions{
		FollowerOnly: followerOnly,
	})
}

func StopSchedulerInConstellation() {
	pro.StopScheduler()
}

func GetCurrentLeaderName() string {
	name, ok := pro.GetLeaderName(&clientConfigLock, nc, js)
	if !ok {
		return ""
	}
	return name
}
