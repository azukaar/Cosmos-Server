package market 

import (
	"net/http"
	"encoding/json"
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
		err := updateCache(w, req)
		if err != nil {
			return
		}
		
		marketGetResult := marketGetResult{
			All: make(map[string]interface{}),
			Showcase: []appDefinition{},
		}

		for _, market := range currentMarketcache {
			results := []appDefinition{}
			for _, app := range market.Results.All {
				// if i < 10 {
					results = append(results, app)
				// } else {
				// 	break
				// }
			}
			marketGetResult.All[market.Name] = results
		}
		
		if len(currentMarketcache) > 0 {
			marketGetResult.Showcase = currentMarketcache[0].Results.Showcase
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