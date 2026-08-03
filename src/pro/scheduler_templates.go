package pro

type NodeIdentity struct {
	DeviceName string
	IP         string
	CosmosNode int
}

func SetNodeIdentityProvider(f func() NodeIdentity) {
	// Pro feature stub: template expansion is only available in Cosmos Pro.
}

func SetMountedStorageProvider(f func() []string) {
	// Pro feature stub.
}
