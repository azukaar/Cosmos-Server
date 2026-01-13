package configapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/azukaar/cosmos-server/src/docker"
	"github.com/azukaar/cosmos-server/src/utils"
)

type DashboardRoute struct {
	Name              string `json:"Name"`
	Description       string `json:"Description,omitempty"`
	Target            string `json:"Target"`
	Mode              string `json:"Mode"`
	Host              string `json:"Host,omitempty"`
	UseHost           bool   `json:"UseHost"`
	UsePathPrefix     bool   `json:"UsePathPrefix"`
	PathPrefix        string `json:"PathPrefix,omitempty"`
	Icon              string `json:"Icon,omitempty"`
	HideFromDashboard bool   `json:"HideFromDashboard"`
	// Container info (for SERVAPP routes)
	ContainerRunning bool   `json:"ContainerRunning"`
	ContainerIcon    string `json:"ContainerIcon,omitempty"`
}

func DashboardApiGet(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	if req.Method != "GET" {
		utils.Error("DashboardApiGet: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}

	isAdmin := utils.IsAdmin(req)
	config := utils.ReadConfigFromFile()

	// Get containers if docker is available
	var containers []struct {
		Names  []string
		State  string
		Labels map[string]string
	}

	containerList, err := docker.ListContainers()
	if err == nil {
		for _, c := range containerList {
			containers = append(containers, struct {
				Names  []string
				State  string
				Labels map[string]string
			}{
				Names:  c.Names,
				State:  c.State,
				Labels: c.Labels,
			})
		}
	}

	// Build dashboard routes
	var dashboardRoutes []DashboardRoute

	for _, route := range config.HTTPConfig.ProxyConfig.Routes {
		// Skip admin-only routes for non-admins
		if !isAdmin && route.AdminOnly {
			continue
		}

		dashRoute := DashboardRoute{
			Name:              route.Name,
			Description:       route.Description,
			Target:            route.Target,
			Mode:              string(route.Mode),
			Host:              route.Host,
			UseHost:           route.UseHost,
			UsePathPrefix:     route.UsePathPrefix,
			PathPrefix:        route.PathPrefix,
			Icon:              route.Icon,
			HideFromDashboard: route.HideFromDashboard,
		}

		// For SERVAPP routes, find matching container
		if route.Mode == "SERVAPP" {
			// Extract container name from target (format: http://containerName:port)
			containerName := extractContainerName(route.Target)

			if containerName != "" {
				for _, c := range containers {
					// Container names include leading slash (e.g., "/containername")
					for _, name := range c.Names {
						if name == "/"+containerName || name == containerName {
							dashRoute.ContainerRunning = c.State == "running"
							if icon, ok := c.Labels["cosmos-icon"]; ok {
								dashRoute.ContainerIcon = icon
							}
							break
						}
					}
				}
			}
		}

		dashboardRoutes = append(dashboardRoutes, dashRoute)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "OK",
		"data":   dashboardRoutes,
	})
}

// extractContainerName extracts the container name from a SERVAPP target URL
// Example: "http://jellyfin:8096" -> "jellyfin"
func extractContainerName(target string) string {
	// Remove protocol prefix
	target = strings.TrimPrefix(target, "http://")
	target = strings.TrimPrefix(target, "https://")

	// Get the host part (before port or path)
	if idx := strings.Index(target, ":"); idx != -1 {
		return target[:idx]
	}
	if idx := strings.Index(target, "/"); idx != -1 {
		return target[:idx]
	}

	return target
}
