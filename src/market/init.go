package market 

import (
	"github.com/azukaar/cosmos-server/src/utils" 
)

func Init() {
	config := utils.GetMainConfig()
	sources := config.MarketConfig.Sources

	inConfig := map[string]bool{
		"cosmos-cloud": true,
	}
	for _, source := range sources {
		inConfig[source.Name] = true
	}
	
	if currentMarketcache == nil {
		currentMarketcache = []marketCacheObject{}
	}

	inCache := map[string]bool{}
	toRemove := []string{}
	for _, cachedMarket := range currentMarketcache {
		inCache[cachedMarket.Name] = true

		if !inConfig[cachedMarket.Name] {
			utils.Log("MarketInit: Removing market " + cachedMarket.Name)
			toRemove = append(toRemove, cachedMarket.Name)
		}
	}

	// remove markets that are not in config
	for _, name := range toRemove {
		for index, cachedMarket := range currentMarketcache {
			if cachedMarket.Name == name {
				currentMarketcache = append(currentMarketcache[:index], currentMarketcache[index+1:]...)
				break
			}
		}
	}

	// prepend the default market
	defaultMarket := utils.MarketSource{
		Url: "https://azukaar.github.io/cosmos-servapps-official/index.json",
		Name: "cosmos-cloud",
	}

	sources = append([]utils.MarketSource{defaultMarket}, sources...)

	for _, marketDef := range sources {
		// add markets that are in config but not in cache
		if !inCache[marketDef.Name] {
			market := marketCacheObject{
				Url: marketDef.Url,
				Name: marketDef.Name,
			}

			currentMarketcache = append(currentMarketcache, market)

			utils.Log("MarketInit: Added market " + market.Name)
		}
	}
}