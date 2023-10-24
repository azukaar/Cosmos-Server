package market

import (
	"encoding/json"
	"github.com/azukaar/cosmos-server/src/utils"
	"os"
)

type MarketConfig struct {
	DefaultMarket utils.MarketSource
	Sources       []utils.MarketSource
}

func Init() {
	config := ReadMarketConfig()
	currentMarketcache = []marketCacheObject{}
	sources := config.Sources

	// Add the default market from the config
	defaultMarket := config.DefaultMarket
	sources = append([]utils.MarketSource{defaultMarket}, sources...)

	for _, marketDef := range sources {
		market := marketCacheObject{
			Url:  marketDef.Url,
			Name: marketDef.Name,
		}
		currentMarketcache = append(currentMarketcache, market)

		utils.Log("MarketInit: Added market " + market.Name)
	}
}

func ReadMarketConfig() MarketConfig {
	configPath := "/config/cosmos.config.json"
	/*Only the market section is needed, but Sources are available to configure as well
		 *
		 *
	    "DefaultMarket": {
	        "Url": "https://custom-default-market.com",
	        "Name": "custom-market"
	    },
	    "Sources": [
	        {
	            "Url": "https://example-repo-1.com",
	            "Name": "example-repo-1"
	        },
	        {
	            "Url": "https://example-repo-2.com",
	            "Name": "example-repo-2"
	        }
	    ]


		 * Add this to the top of the config to change the defaults.
	*/
	file, err := os.Open(configPath)
	if err != nil {
		// Handle the error (e.g., log it or provide a default config)
		// You might want to provide a default config here in case the file is missing or invalid.
		defaultConfig := MarketConfig{
			DefaultMarket: utils.MarketSource{
				Url:  "https://cosmos-cloud.io/repository",
				Name: "cosmos-cloud",
			},
			Sources: []utils.MarketSource{},
		}
		return defaultConfig
	}
	defer file.Close()

	var config MarketConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		// Handle the error (e.g., log it or provide a default config)
		// You might want to provide a default config here in case the file is invalid.
		defaultConfig := MarketConfig{
			DefaultMarket: utils.MarketSource{
				Url:  "https://cosmos-cloud.io/repository",
				Name: "cosmos-cloud",
			},
			Sources: []utils.MarketSource{},
		}
		return defaultConfig
	}

	return config
}
