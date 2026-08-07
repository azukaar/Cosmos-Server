package pro

import (
	"net/http"
	
	"github.com/azukaar/cosmos-server/src/utils"
)

type CreateGroupRequest struct {
	Name        string             `json:"name" validate:"required"`
	Permissions []utils.Permission `json:"permissions"`
}

type UpdateGroupRequest struct {
	Name        *string            `json:"name,omitempty"`
	Permissions []utils.Permission `json:"permissions,omitempty"`
}

func GroupsRoute(w http.ResponseWriter, req *http.Request) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
	return
}

func GroupsIdRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "PUT" {
		updateGroup(w, req)
	} else if req.Method == "DELETE" {
		deleteGroup(w, req)
	} else {
		utils.Error("GroupsIdRoute: Method not allowed "+req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
	}
}

// listGroups godoc
// @Summary List all groups
// @Description Returns all custom role groups (role ID >= 3)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.APIResponse
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Router /api/groups [get]
func listGroups(w http.ResponseWriter, req *http.Request) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
	return
}

// createGroup godoc
// @Summary Create a new group
// @Description Creates a new custom role group with name and permissions (Pro feature)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateGroupRequest true "Group creation details"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/groups [post]
func createGroup(w http.ResponseWriter, req *http.Request) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
	return
}

// updateGroup godoc
// @Summary Update a group
// @Description Updates an existing group's name or permissions (Pro feature)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Param request body UpdateGroupRequest true "Fields to update"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/groups/{id} [put]
func updateGroup(w http.ResponseWriter, req *http.Request) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
	return
}

// deleteGroup godoc
// @Summary Delete a group
// @Description Deletes a custom role group. Fails if users are still assigned to it. (Pro feature)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Success 200 {object} utils.APIResponse
// @Failure 400 {object} utils.HTTPErrorResult
// @Failure 401 {object} utils.HTTPErrorResult
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 404 {object} utils.HTTPErrorResult
// @Failure 409 {object} utils.HTTPErrorResult
// @Router /api/groups/{id} [delete]
func deleteGroup(w http.ResponseWriter, req *http.Request) {
	utils.Error("This is a pro and is not currently available on your server. Please upgrade to Cosmos Pro to access this feature.", nil)
	utils.HTTPError(w, "This feature is only available in Cosmos Pro", http.StatusForbidden, "PRO001")
	return
}
