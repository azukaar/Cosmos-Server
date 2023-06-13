package market 

import (
	"github.com/azukaar/cosmos-server/src/utils" 
)

func Init() {
	config := utils.GetMainConfig()
	currentMarketcache = []marketCacheObject{}
	sources := config.MarketConfig.Sources

	// prepend the default market
	defaultMarket := utils.MarketSource{
		Url: "https://cosmos-cloud.io/repository",
		Name: "cosmos-cloud",
	}
	sources = append([]utils.MarketSource{defaultMarket}, sources...)

	for _, marketDef := range sources {
		market := marketCacheObject{
			Url: marketDef.Url,
			Name: marketDef.Name,
		}
		currentMarketcache = append(currentMarketcache, market)

		utils.Log("MarketInit: Added market " + market.Name)
	}
}