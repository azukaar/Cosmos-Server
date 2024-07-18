package market

import (
	"net/http"
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils" 
	"time"
)
type appDefinition struct {
	Name                   string                `json:"name"`
	Description            string                `json:"description"`
	Url                    string                `json:"url"`
	LongDescription        string                `json:"longDescription"`
	Translation            *translationLanguages `json:"translation"`
	Tags                   []string              `json:"tags"`
	Repository             string                `json:"repository"`
	Image                  string                `json:"image"`
	Screenshots            []string              `json:"screenshots"`
	Icon                   string                `json:"icon"`
	Compose                string                `json:"compose"`
	SupportedArchitectures []string              `json:"supported_architectures"`
}

type translationLanguages struct {
	En		*appDefinition `json:"en"`
	De 		*appDefinition `json:"de"`
	DeCH	*appDefinition `json:"de-CH"`
}

type translationFields struct {
	Description     string `json:"description"`
	LongDescription string `json:"longDescription"`
}

type marketDefinition struct {
	Showcase []appDefinition `json:"showcase"`
	All      []appDefinition `json:"all"`
	Source   string          `json:"source"`
}

type marketCacheObject struct {
	Url        string           `json:"url"`
	Name       string           `json:"name"`
	LastUpdate time.Time        `json:"lastUpdate"`
	Results    marketDefinition `json:"results"`
}

var currentMarketcache []marketCacheObject

func GetCachedMarket() []marketCacheObject {
	return currentMarketcache
}

func updateCache(w http.ResponseWriter, req *http.Request) error {
	for index, cachedMarket := range currentMarketcache {
		if cachedMarket.LastUpdate.Add(time.Hour * 12).Before(time.Now()) {
			utils.Log("MarketUpdate: Updating market " + cachedMarket.Name)

			// fetch market.url
			resp, err := http.Get(cachedMarket.Url)
			if err != nil {
				utils.Error("MarketUpdate: Error while fetching market"+cachedMarket.Url, err)
				// utils.HTTPError(w, "Market Get Error " + cachedMarket.Url, http.StatusInternalServerError, "MK001")
				continue
			}

			defer resp.Body.Close()

			// parse body
			var result marketDefinition
			err = json.NewDecoder(resp.Body).Decode(&result)

			if err != nil {
				utils.Error("MarketUpdate: Error while parsing market"+cachedMarket.Url, err)
				// utils.HTTPError(w, "Market Get Error " + cachedMarket.Url, http.StatusInternalServerError, "MK003")
				continue
			}

			result.Source = cachedMarket.Url
			if cachedMarket.Name != "cosmos-cloud" {
				result.Showcase = []appDefinition{}
			}

			cachedMarket.Results = result
			cachedMarket.LastUpdate = time.Now()

			utils.TriggerEvent(
				"cosmos.market.update",
				"Market updated",
				"success",
				"",
				map[string]interface{}{
					"market":       cachedMarket.Name,
					"numberOfApps": len(result.All),
				})

			utils.Log("MarketUpdate: Updated market " + result.Source + " with " + string(len(result.All)) + " results")

			// save to cache
			currentMarketcache[index] = cachedMarket
		}
	}

	return nil
}
