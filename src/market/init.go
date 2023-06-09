package market 

import (
	"github.com/azukaar/cosmos-server/src/utils" 
)

func Init() {
	config := utils.GetMainConfig()
	currentMarketcache = []marketCacheObject{}

	for _, marketDef := range config.MarketConfig.Sources {
		market := marketCacheObject{
			Url: marketDef.Url,
			Name: marketDef.Name,
		}
		currentMarketcache = append(currentMarketcache, market)

		utils.Log("MarketInit: Added market " + market.Name)
	}
}