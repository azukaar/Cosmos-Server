package metrics 

import (
	"strconv"
	"strings"
	"time"	
	"os"
	"context"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
  "github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/common"
	
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/docker"
)

func GetSystemMetrics() {
	utils.Debug("Metrics - System")

	ctx := context.Background()

	// redirect docker monitoring
	if os.Getenv("HOSTNAME") != "" {
		// check if path /mnt/host exist
		if _, err := os.Stat("/mnt/host"); os.IsNotExist(err) {
			utils.Error("Metrics - Cannot start monitoring the server if you don't mount /mnt/host to /. Check the documentation for more information.", nil)
			return
		} else {
			utils.Log("Metrics - Monitoring the server at /mnt/host")

			ctx = context.WithValue(context.Background(), 
				common.EnvKey, common.EnvMap{
					common.HostProcEnvKey: "/mnt/host/proc",
					common.HostSysEnvKey: "/mnt/host/sys",
					common.HostEtcEnvKey: "/mnt/host/etc",
					common.HostVarEnvKey: "/mnt/host/var",
					common.HostRunEnvKey: "/mnt/host/run",
					common.HostDevEnvKey: "/mnt/host/dev",
					common.HostRootEnvKey: "/mnt/host/",
				},
			)
		}
	}


	// Get CPU Usage
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		utils.Error("Metrics - Error fetching CPU usage:", err)
		return
	}
	if len(cpuPercent) > 0 {
		for i, v := range cpuPercent {
			PushSetMetric("system.cpu." + strconv.Itoa(i), int(v), DataDef{
				Max: 100,
				Period: time.Second * 30,
				Label: "CPU " + strconv.Itoa(i),
				AggloType: "avg",
			})
		}
	}

	// You can get detailed per-CPU stats with:
	// cpuPercents, _ := cpu.Percent(0, true)

	// Get RAM Usage
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		utils.Error("Metrics - Error fetching RAM usage:", err)
		return
	}
	PushSetMetric("system.ram", int(memInfo.Used), DataDef{
		Max: memInfo.Total,
		Period: time.Second * 30,
		Label: "RAM",
		AggloType: "avg",
	})
	
	// Get Network Usage
	netIO, err := net.IOCountersWithContext(ctx, false)
	
	netIOTest, _ := net.IOCounters(true)
	for _, v := range netIOTest {
		utils.Debug("Metrics - Network " + v.Name + " : " + strconv.Itoa(int(v.BytesRecv)) + " / " + strconv.Itoa(int(v.BytesSent)) + " / " + strconv.Itoa(int(v.Errin + v.Errout)) + " / " + strconv.Itoa(int(v.Dropin + v.Dropout)))
	}

	PushSetMetric("system.netRx", int(netIO[0].BytesRecv), DataDef{
		Max: 0,
		Period: time.Second * 30,
		Label: "Network Received",
		AggloType: "avg",
	})

	PushSetMetric("system.netTx", int(netIO[0].BytesSent), DataDef{
		Max: 0,
		Period: time.Second * 30,
		Label: "Network Sent",
		AggloType: "avg",
	})

	PushSetMetric("system.netErr", int(netIO[0].Errin + netIO[0].Errout), DataDef{
		Max: 0,
		Period: time.Second * 30,
		Label: "Network Errors",
		AggloType: "avg",
	})

	PushSetMetric("system.netDrop", int(netIO[0].Dropin + netIO[0].Dropout), DataDef{
		Max: 0,
		Period: time.Second * 30,
		Label: "Network Drops",
		AggloType: "avg",
	})

	// docker stats
	dockerStats, err := docker.StatsAll()
	if err != nil {
		utils.Error("Metrics - Error fetching Docker stats:", err)
		return
	}

	for _, ds := range dockerStats {
		PushSetMetric("system.docker.cpu." + ds.Name, int(ds.CPUUsage), DataDef{
			Max: 100,
			Period: time.Second * 30,
			Label: "Docker CPU " + ds.Name,
			AggloType: "avg",
		})
		PushSetMetric("system.docker.ram." + ds.Name, int(ds.MemUsage), DataDef{
			Max: 100,
			Period: time.Second * 30,
			Label: "Docker RAM " + ds.Name,
			AggloType: "avg",
		})
		PushSetMetric("system.docker.netRx." + ds.Name, int(ds.NetworkRx), DataDef{
			Max: 0,
			Period: time.Second * 30,
			Label: "Docker Network Received " + ds.Name,
			AggloType: "avg",
		})
		PushSetMetric("system.docker.netTx." + ds.Name, int(ds.NetworkTx), DataDef{
			Max: 0,
			Period: time.Second * 30,
			Label: "Docker Network Sent " + ds.Name,
			AggloType: "avg",
		})
	}

	// Get Disk Usage
  parts, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		utils.Error("Metrics - Error fetching Disk usage:", err)
		return
	}

  for _, part := range parts {
		if strings.HasPrefix(part.Mountpoint, "/dev") || (strings.HasPrefix(part.Mountpoint, "/mnt") && !strings.HasPrefix(part.Mountpoint, "/mnt/host")) {
			realMount := part.Mountpoint
			
			if os.Getenv("HOSTANME") != "" {
				realMount = "/mnt/host" + part.Mountpoint
			}

			u, err := disk.Usage(realMount)
			if err != nil {
				utils.Error("Metrics - Error fetching Disk usage for " + realMount + " : ", err)
			} else {
				PushSetMetric("system.disk." + part.Mountpoint, int(u.Used), DataDef{
					Max: u.Total,
					Period: time.Second * 120,
					Label: "Disk " + part.Mountpoint,
				})
			}
		}
	}
}