package proxy

import (
	"net/http"

	"github.com/azukaar/cosmos-server/src/metrics"
)

var botUserAgents = []string{
	"360Spider", "acapbot", "acoonbot", "ahrefs", "alexibot", "asterias", "attackbot", "backdorbot", "becomebot", "binlar",
	"blackwidow", "blekkobot", "blexbot", "blowfish", "bullseye", "bunnys", "butterfly", "careerbot", "casper", "checkpriv",
	"cheesebot", "cherrypick", "chinaclaw", "choppy", "clshttp", "cmsworld", "copernic", "copyrightcheck", "cosmos", "crescent",
	"cy_cho", "datacha", "demon", "diavol", "discobot", "dittospyder", "dotbot", "dotnetdotcom", "dumbot", "emailcollector",
	"emailsiphon", "emailwolf", "exabot", "extract", "eyenetie", "feedfinder", "flaming", "flashget", "flicky", "foobot",
	"g00g1e", "getright", "gigabot", "go-ahead-got", "gozilla", "grabnet", "grafula", "harvest", "heritrix", "httrack",
	"icarus6j", "jetbot", "jetcar", "jikespider", "kmccrew", "leechftp", "libweb", "linkextractor", "linkscan", "linkwalker",
	"loader", "masscan", "miner", "majestic", "mechanize", "mj12bot", "morfeus", "moveoverbot", "netmechanic", "netspider",
	"nicerspro", "nikto", "ninja", "nutch", "octopus", "pagegrabber", "planetwork", "postrank", "proximic", "purebot",
	"pycurl", "python", "queryn", "queryseeker", "radian6", "radiation", "realdownload", "rogerbot", "scooter", "seekerspider",
	"semalt", "siclab", "sindice", "sistrix", "sitebot", "siteexplorer", "sitesnagger", "skygrid", "smartdownload", "snoopy",
	"sosospider", "spankbot", "spbot", "sqlmap", "stackrambler", "stripper", "sucker", "surftbot", "sux0r", "suzukacz",
	"suzuran", "takeout", "teleport", "telesoft", "true_robots", "turingos", "turnit", "vampire", "vikspider", "voideye",
	"webleacher", "webreaper", "webstripper", "webvac", "webviewer", "webwhacker", "winhttp", "wwwoffle", "woxbot",
	"xaldon", "xxxyy", "yamanalab", "yioopbot", "youda", "zeus", "zmeu", "zune", "zyborg",
}

// botDetectionMiddleware checks if the User-Agent is a known bot
func BotDetectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.UserAgent()
		
		if userAgent == "" {
			go metrics.PushShieldMetrics("bots")
			http.Error(w, "Access denied: Bots are not allowed.", http.StatusForbidden)
			return
		}

		for _, botUserAgent := range botUserAgents {
			if userAgent == botUserAgent {
			go metrics.PushShieldMetrics("bots")
				http.Error(w, "Access denied: Bots are not allowed.", http.StatusForbidden)
				return
			}
		}

		// If no bot user agent is detected, pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}