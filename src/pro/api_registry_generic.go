// Community build stub of the Cosmos Pro feature set; handlers answer PRO001.

package pro

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/nats-io/nats.go"
	"net/http"
	"sync"
)

// RegistryGenericPackagesRoute godoc
// @Summary List the packages of a generic or pypi registry
// @Description Returns every package with its versions and their files (Pro feature)
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries/{name}/packages [get]
func RegistryGenericPackagesRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistryGenericPackageIdRoute godoc
// @Summary Get or delete one generic or pypi package
// @Description GET returns the package with its versions and files; DELETE removes the
// @Description package and every version (the stored files are reclaimed by the next
// @Description GC pass) (Pro feature)
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Param package path string true "Package name"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries/{name}/packages/{package} [get]
func RegistryGenericPackageIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistryGenericVersionsRoute godoc
// @Summary Upload files into a generic package version
// @Description Stores one or more files under a version. Accepts a multipart form (every
// @Description file part is stored under its own filename) or a single raw file as the
// @Description request body, in which case the filename query parameter is required.
// @Description Query parameters: version (default: a UTC timestamp), filename. A file
// @Description that already exists in the version is refused (409): upload a new version
// @Description or delete it first. Accepts a Cosmos token with the Resources permission
// @Description OR a registry deploy token with push scope on an access that exposes this
// @Description registry (Pro feature).
// @Tags registry
// @Accept octet-stream
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Param package path string true "Package name"
// @Param version query string false "Version (default: UTC timestamp)"
// @Param filename query string false "Filename, required for a raw body"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries/{name}/packages/{package}/versions [post]
func RegistryGenericVersionsRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistryGenericVersionIdRoute godoc
// @Summary Delete one generic or pypi package version
// @Description Removes the version and its file entries; the stored files are reclaimed
// @Description by the next GC pass. "latest" is re-pointed at the newest remaining
// @Description version (Pro feature).
// @Tags registry
// @Produce json
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Param package path string true "Package name"
// @Param version path string true "Version"
// @Success 200 {object} utils.APIResponse
// @Router /api/constellation/registries/{name}/packages/{package}/versions/{version} [delete]
func RegistryGenericVersionIdRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}

// RegistryGenericFileRoute godoc
// @Summary Download or delete one file of a generic or pypi package version
// @Description GET streams the file (the version may be "latest"); accepts a Cosmos token
// @Description with the Resources read permission OR a registry deploy token with pull
// @Description scope. DELETE removes the file entry — and the version, when it was its
// @Description last file; the stored bytes are reclaimed by the next GC pass (Pro feature).
// @Tags registry
// @Produce octet-stream
// @Security BearerAuth
// @Param name path string true "Registry name"
// @Param package path string true "Package name"
// @Param version path string true "Version, or latest"
// @Param file path string true "Filename"
// @Success 200 {file} binary
// @Router /api/constellation/registries/{name}/packages/{package}/versions/{version}/files/{file} [get]
func RegistryGenericFileRoute(w http.ResponseWriter, req *http.Request, lock *sync.RWMutex, js nats.JetStreamContext) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
}
