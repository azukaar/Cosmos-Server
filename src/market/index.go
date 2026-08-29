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

// MarketGet godoc
// @Summary Get the app marketplace listings
// @Description Returns all marketplace sources and their applications, plus the showcase from cosmos-cloud.
// @Tags Market
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} utils.HTTPErrorResult
// @Failure 500 {object} utils.HTTPErrorResult
// @Router /api/markets [get]
func MarketGet(w http.ResponseWriter, req *http.Request) {
	if utils.CheckPermissions(w, req, utils.PERM_LOGIN) != nil {
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
				// Each app gets its own batch of freshly-generated secret
				// encryption keys (64-hex) so installer templates can use
				// {Secrets.0}, {Secrets.1}, ... Every index is a different
				// random value, and the batch is regenerated on each fetch
				// (never persisted). Only as many keys as the template
				// references are generated (max index + 1, min 1).
				n := maxSecretsIndex(app.Compose) + 1
				if n < 1 {
					n = 1
				}
				app.Secrets = utils.GenerateSecrets(n)
				results = append(results, app)
			}
			marketGetResult.All[market.Name] = results
		}
		
		if len(currentMarketcache) > 0 {
			for _, market := range currentMarketcache {
				if market.Name == "cosmos-cloud" {
					showcase := []appDefinition{}
					for _, app := range market.Results.Showcase {
						n := maxSecretsIndex(app.Compose) + 1
						if n < 1 {
							n = 1
						}
						app.Secrets = utils.GenerateSecrets(n)
						showcase = append(showcase, app)
					}
					marketGetResult.Showcase = showcase
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