package market 

import (
	"net/http"
	"encoding/json"
	"fmt"
	"github.com/azukaar/cosmos-server/src/utils" 
)

type marketGetResult struct {
	Showcase []appDefinition `json:"showcase"`
	All map[string]interface{} `json:"all"`
}

func MarketGet(w http.ResponseWriter, req *http.Request) {
	if utils.LoggedInOnly(w, req) != nil {
		return
	}

	if(req.Method == "GET") {
		config := utils.GetMainConfig()
		configSourcesList := config.MarketConfig.Sources
		configSources := map[string]bool{
			"cosmos-cloud": true,
		}
		for _, source := range configSourcesList {
			configSources[source.Name] = true
		}

		utils.Debug(fmt.Sprintf("MarketGet: Config sources: %v", configSources))

		Init()

		err := updateCache(w, req)
		if err != nil {
			utils.Error("MarketGet: Error while updating cache", err)
			utils.HTTPError(w, "Error while updating cache", http.StatusInternalServerError, "MK002")
			return
		}
		
		marketGetResult := marketGetResult{
			All: make(map[string]interface{}),
			Showcase: []appDefinition{},
		}

		for _, market := range currentMarketcache {
			if !configSources[market.Name] {
				continue
			}
			utils.Debug(fmt.Sprintf("MarketGet: Adding market %v", market.Name))
			results := []appDefinition{}
			for _, app := range market.Results.All {
				results = append(results, app)
			}
			marketGetResult.All[market.Name] = results
		}
		
		if len(currentMarketcache) > 0 {
			for _, market := range currentMarketcache {
				if market.Name == "cosmos-cloud" {
					marketGetResult.Showcase = market.Results.Showcase
				}
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "OK",
			"data": marketGetResult,
		})
	} else {
		utils.Error("MarketGet: Method not allowed" + req.Method, nil)
		utils.HTTPError(w, "Method not allowed", http.StatusMethodNotAllowed, "HTTP001")
		return
	}
}