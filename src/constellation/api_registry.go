package constellation

import (
	"net/http"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/azukaar/cosmos-server/src/pro"
)

// Thin bridges keeping `pro` independent of `constellation`; js is read at call time so reconnects are picked up.

func RegistryRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryRoute(w, req, &clientConfigLock, js)
}

func RegistryIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryIdRoute(w, req, &clientConfigLock, js)
}

func RegistrySettingsRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistrySettingsRoute(w, req, &clientConfigLock, js)
}

func RegistryGCRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryGCRoute(w, req, &clientConfigLock, js)
}

func RegistryStaticSitesRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryStaticSitesRoute(w, req, &clientConfigLock, js)
}

func RegistryStaticSiteIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryStaticSiteIdRoute(w, req, &clientConfigLock, js)
}

func RegistryStaticVersionsRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryStaticVersionsRoute(w, req, &clientConfigLock, js)
}

func RegistryStaticVersionIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryStaticVersionIdRoute(w, req, &clientConfigLock, js)
}

func RegistryStaticActivateRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryStaticActivateRoute(w, req, &clientConfigLock, js)
}

func RegistryStaticVersionDownloadRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryStaticVersionDownloadRoute(w, req, &clientConfigLock, js)
}

func RegistryGenericPackagesRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryGenericPackagesRoute(w, req, &clientConfigLock, js)
}

func RegistryGenericPackageIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryGenericPackageIdRoute(w, req, &clientConfigLock, js)
}

func RegistryGenericVersionsRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryGenericVersionsRoute(w, req, &clientConfigLock, js)
}

func RegistryGenericVersionIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryGenericVersionIdRoute(w, req, &clientConfigLock, js)
}

func RegistryGenericFileRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryGenericFileRoute(w, req, &clientConfigLock, js)
}

func RegistryAccessRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryAccessRoute(w, req, &clientConfigLock, js)
}

func RegistryAccessIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryAccessIdRoute(w, req, &clientConfigLock, js)
}

func RegistryAccessSettingsRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryAccessSettingsRoute(w, req, &clientConfigLock, js)
}

func RegistryAccessTokensRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryAccessTokensRoute(w, req, &clientConfigLock, js)
}

func RegistryAccessTokenIdRoute(w http.ResponseWriter, req *http.Request) {
	pro.RegistryAccessTokenIdRoute(w, req, &clientConfigLock, js)
}

// RegistryList is best-effort: a KV failure returns nil, treated as "unavailable", never "no registries".
func RegistryList() []pro.RegistryStatus {
	list, err := pro.ListRegistriesWithStatus(&clientConfigLock, js)
	if err != nil {
		return nil
	}
	return list
}

func RegistryAccessList() []pro.RegistryAccessStatus {
	list, err := pro.ListRegistryAccessesWithStatus(&clientConfigLock, js)
	if err != nil {
		return nil
	}
	return list
}

// RegistryClusterHandles hands pro the current cluster handles, read at call time.
func RegistryClusterHandles() (*sync.RWMutex, nats.JetStreamContext, *nats.Conn) {
	return &clientConfigLock, js, nc
}

// RegistryNodeInfo projects this device onto the registry serving paths: placement tags and the node-to-node auth key.
func RegistryNodeInfo() pro.RegistryNodeInfo {
	device, err := GetCurrentDevice()
	if err != nil {
		return pro.RegistryNodeInfo{}
	}
	return pro.RegistryNodeInfo{
		DeviceName: device.DeviceName,
		IP:         device.IP,
		Tags:       device.Tags,
		APIKey:     device.APIKey,
	}
}
