package pro

type NodeResources struct {
	CPUPercent   float64
	RAMPercent   float64
	MonitoringOn bool
}

func StartResourceSampler() {
	// Pro feature stub: cluster monitoring is only available in Cosmos Pro.
}

func StopResourceSampler() {
	// Pro feature stub.
}

func GetCurrentResources() NodeResources {
	return NodeResources{}
}
