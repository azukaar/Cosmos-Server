package storage

import (
	"encoding/json"
	"net/http"

	"github.com/azukaar/cosmos-server/src/utils"
)

func ListDisksRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		disks, err := ListDisks()
		if err != nil {
			utils.Error("ListDisksRoute: " + err.Error(), nil)
			utils.HTTPError(w, "Internal server error", http.StatusInternalServerError, "STO001")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   disks,
		})
	} else {
		utils.Error("ListDisksRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}

func ListMountsRoute(w http.ResponseWriter, req *http.Request) {
	if utils.AdminOnly(w, req) != nil {
		return
	}

	if req.Method == "GET" {
		disks, err := ListMounts()
		if err != nil {
			utils.Error("ListMountsRoute: " + err.Error(), nil)
			utils.HTTPError(w, "Internal server error", http.StatusInternalServerError, "STO001")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data":   disks,
		})
	} else {
		utils.Error("ListMountsRoute: Method not allowed " + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}