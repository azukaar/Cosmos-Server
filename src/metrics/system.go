package metrics

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/common"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/storage"
	"github.com/azukaar/cosmos-server/src/utils"
)

func GetSystemMetrics() {
	utils.Debug("Metrics - System")

	ctx := context.Background()

	// redirect docker monitoring
	if utils.IsInsideContainer {
		// check if path /mnt/host exist
		if _, err := os.Stat("/mnt/host"); os.IsNotExist(err) {
			utils.Error("Metrics - Cannot start monitoring the server if you don't mount /mnt/host to /. Check the documentation for more information.", nil)
			return
		} else {
			utils.Log("Metrics - Monitoring the server at /mnt/host")

			ctx = context.WithValue(context.Background(),
				common.EnvKey, common.EnvMap{
					common.HostProcEnvKey: "/mnt/host/proc",
					common.HostSysEnvKey:  "/mnt/host/sys",
					common.HostEtcEnvKey:  "/mnt/host/etc",
					common.HostVarEnvKey:  "/mnt/host/var",
					common.HostRunEnvKey:  "/mnt/host/run",
					common.HostDevEnvKey:  "/mnt/host/dev",
					common.HostRootEnvKey: "/mnt/host/",
				},
			)
		}
	}

	// Get CPU Usage
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		utils.Error("Metrics - Error fetching CPU usage:", err)
	} else {
		if len(cpuPercent) > 0 {
			for i, v := range cpuPercent {
				PushSetMetric("system.cpu."+strconv.Itoa(i), int(v), DataDef{
					Max:       100,
					Period:    time.Second * 30,
					Label:     "CPU " + strconv.Itoa(i),
					AggloType: "avg",
					Unit:      "%",
				})
			}
		}
	}

	// You can get detailed per-CPU stats with:
	// cpuPercents, _ := cpu.Percent(0, true)

	// Get RAM Usage
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		utils.Error("Metrics - Error fetching RAM usage:", err)
	} else {
		PushSetMetric("system.ram", int(memInfo.Used), DataDef{
			Max:       memInfo.Total,
			Period:    time.Second * 30,
			Label:     "RAM",
			AggloType: "avg",
			Unit:      "B",
		})
	}

	// Get Network Usage
	netIO, err := net.IOCountersWithContext(ctx, false)

	// netIOTest, _ := net.IOCountersWithContext(ctx, true)
	// for _, v := range netIOTest {
	// 	utils.Debug("Metrics - Network " + v.Name + " : " + strconv.Itoa(int(v.BytesRecv)) + " / " + strconv.Itoa(int(v.BytesSent)) + " / " + strconv.Itoa(int(v.Errin + v.Errout)) + " / " + strconv.Itoa(int(v.Dropin + v.Dropout)))
	// }

	if err != nil {
		utils.Error("Metrics - Error fetching Network usage:", err)
	} else {
		PushSetMetric("system.netRx", int(netIO[0].BytesRecv), DataDef{
			Max:          0,
			Period:       time.Second * 30,
			Label:        "Network Received",
			SetOperation: "max",
			AggloType:    "sum",
			Decumulate:   true,
			Unit:         "B",
		})

		PushSetMetric("system.netTx", int(netIO[0].BytesSent), DataDef{
			Max:          0,
			Period:       time.Second * 30,
			Label:        "Network Sent",
			SetOperation: "max",
			AggloType:    "sum",
			Decumulate:   true,
			Unit:         "B",
		})

		PushSetMetric("system.netErr", int(netIO[0].Errin+netIO[0].Errout), DataDef{
			Max:          0,
			Period:       time.Second * 30,
			Label:        "Network Errors",
			SetOperation: "max",
			AggloType:    "sum",
			Decumulate:   true,
		})

		PushSetMetric("system.netDrop", int(netIO[0].Dropin+netIO[0].Dropout), DataDef{
			Max:          0,
			Period:       time.Second * 30,
			Label:        "Network Drops",
			SetOperation: "max",
			AggloType:    "sum",
			Decumulate:   true,
		})
	}

	// Get Disk Usage
	parts, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		utils.Error("Metrics - Error fetching Disk usage:", err)
	} else {
		for _, part := range parts {
			if strings.HasPrefix(part.Mountpoint, "/dev") ||
				(strings.HasPrefix(part.Mountpoint, "/mnt") && !strings.HasPrefix(part.Mountpoint, "/mnt/host") ||
					part.Mountpoint == "/") {

				if part.Mountpoint == "/dev" || strings.HasPrefix(part.Mountpoint, "/dev/shm") || strings.HasPrefix(part.Mountpoint, "/dev/pts") {
					continue
				}

				realMount := part.Mountpoint
				mountKey := strings.Replace(part.Mountpoint, ".", "_", -1)

				if utils.IsInsideContainer {
					realMount = "/mnt/host" + part.Mountpoint
				}

				u, err := disk.Usage(realMount)
				if err != nil {
					utils.Error("Metrics - Error fetching Disk usage for "+realMount+" : ", err)
				} else {
					PushSetMetric("system.disk."+mountKey, int(u.Used), DataDef{
						Max:    u.Total,
						Period: time.Second * 120,
						Label:  "Disk " + part.Mountpoint,
						Unit:   "B",
						Object: "disk@" + part.Mountpoint,
					})
				}
			}
		}
	}

	// Temperature
	temps, err := host.SensorsTemperatures()
	if err != nil {
		utils.Error("Metrics - Error fetching Temperature:", err)
	} else {
		avgTemp := 0
		avgTempCount := 0

		for _, temp := range temps {
			utils.Debug("Metrics - Temperature " + temp.SensorKey + " : " + strconv.Itoa(int(temp.Temperature)))
			if temp.Temperature > 0 {
				avgTemp += int(temp.Temperature)
				avgTempCount++

				PushSetMetric("system.temp."+temp.SensorKey, int(temp.Temperature), DataDef{
					Max:    0,
					Period: time.Second * 30,
					Label:  "Temperature " + temp.SensorKey,
					Unit:   "°C",
				})
			}
		}

		if avgTempCount > 0 {
			PushSetMetric("system.temp.all", avgTemp/avgTempCount, DataDef{
				Max:    0,
				Period: time.Second * 30,
				Label:  "Temperature - All",
			})
		}
	}

	// docker stats
	dockerStats, err := docker.StatsAll()
	if err != nil {
		utils.Error("Metrics - Error fetching Docker stats:", err)
	} else {
		for _, ds := range dockerStats {
			containerName := strings.Replace(ds.Name, ".", "_", -1)

			PushSetMetric("system.docker.cpu."+containerName, int(ds.CPUUsage), DataDef{
				Period:    time.Second * 30,
				Label:     "Docker CPU " + containerName,
				AggloType: "avg",
				Scale:     100,
				Unit:      "%",
				Object:    "container@" + containerName,
			})
			PushSetMetric("system.docker.ram."+containerName, int(ds.MemUsage), DataDef{
				Max:       memInfo.Total,
				Period:    time.Second * 30,
				Label:     "Docker RAM " + containerName,
				AggloType: "avg",
				Unit:      "B",
				Object:    "container@" + containerName,
			})
			PushSetMetric("system.docker.netRx."+containerName, int(ds.NetworkRx), DataDef{
				Max:           0,
				Period:        time.Second * 30,
				Label:         "Docker Network Received " + containerName,
				SetOperation:  "max",
				AggloType:     "sum",
				DecumulatePos: true,
				Unit:          "B",
				Object:        "container@" + containerName,
			})
			PushSetMetric("system.docker.netTx."+containerName, int(ds.NetworkTx), DataDef{
				Max:           0,
				Period:        time.Second * 30,
				Label:         "Docker Network Sent " + containerName,
				SetOperation:  "max",
				AggloType:     "sum",
				DecumulatePos: true,
				Unit:          "B",
				Object:        "container@" + containerName,
			})
		}
	}

	// Disk health
	disks, err := storage.ListDisks()
	if err != nil {
		utils.Error("Metrics - Error fetching Disk health:", err)
	} else {
		for _, disk := range disks {
			diskName := strings.Replace(disk.Name, ".", "_", -1)

			PushSetMetric("system.disk-health.temperature."+diskName, int(disk.SMART.Temperature), DataDef{
				Period:    time.Second * 60,
				Label:     "Disk Temperature " + diskName,
				AggloType: "avg",
				Unit:      "°C",
				Object:    "disk@" + diskName,
			})
		}
	}

	// RClone stats
	rcloneStats, err := storage.RCloneStats()

	if err != nil {
		utils.Error("Metrics - Error fetching RClone stats:", err)
	} else {
		PushSetMetric("system.rclone.all.bytes", int(rcloneStats.Bytes), DataDef{
			Period:    time.Second * 30,
			Label:     "Remote Storage Bytes Transfered",
			AggloType: "sum",
			SetOperation: "max",
			Unit:      "B",
			Decumulate:   true,
		})
		PushSetMetric("system.rclone.all.errors", int(rcloneStats.Errors), DataDef{
			Period:    time.Second * 30,
			Label:     "Remote Storage Errors",
			AggloType: "sum",
			SetOperation: "max",
			Unit:      "",
			Decumulate:   true,
		})
	}
}
