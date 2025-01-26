package proxy

import (
	"errors"
	"net/http"
	"strconv"
	"fmt"
	"net/url"
	"crypto/sha256"
	"encoding/base64"

	"github.com/azukaar/cosmos-server/src/utils"
)

func performLogin(w http.ResponseWriter, req *http.Request, route utils.ProxyRouteConfig) error {
	config := utils.GetMainConfig()
	pathHash := sha256.Sum256([]byte(req.URL.Path + config.HTTPConfig.AuthPrivateKey[32:64]))
	// Take first 16 bytes of hash and encode to base64url for shorter string
	hashStr := base64.RawURLEncoding.EncodeToString(pathHash[:16])
	// Combine hash and original path with comma
	state := fmt.Sprintf("%s,,%s", hashStr, url.QueryEscape(req.URL.Path))

	// Get proxy client info if this is a proxy route
	if route.AuthEnabled {
			client := utils.GetProxyOIDCredentials(route, false)
			mainDomain := utils.GetMainConfig().HTTPConfig.Hostname
			if utils.IsHTTPS {
					mainDomain = "https://" + mainDomain
			} else {
					mainDomain = "http://" + mainDomain
			}
			
			//TODO: State should be a random string
			authURL := fmt.Sprintf("%s/cosmos-ui/openid?"+
					"response_type=code&"+
					"client_id=%s&"+
					"redirect_uri=%s&"+
					"scope=openid&"+
					"state=%s",
					mainDomain,
					client.ID,
					url.QueryEscape(client.RedirectURIs[0]),
					url.QueryEscape(state),
			)
			
			http.Redirect(w, req, authURL, http.StatusFound)
			return errors.New("User not logged in, redirecting to OpenID")
	}

	return nil
}

func LoggedInOnlyWithRedirect(w http.ResponseWriter, req *http.Request, route utils.ProxyRouteConfig) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	// mfa, _ := strconv.Atoi(req.Header.Get("x-cosmos-mfa"))
	isUserLoggedIn := role >= 0

	if !isUserLoggedIn || userNickname == "" {
		utils.Error("App gate: User is not logged in", nil)
		return performLogin(w, req, route)
	}

	return nil
}

func AdminOnlyWithRedirect(w http.ResponseWriter, req *http.Request, route utils.ProxyRouteConfig) error {
	userNickname := req.Header.Get("x-cosmos-user")
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	userRole, _ := strconv.Atoi(req.Header.Get("x-cosmos-user-role"))
	
	isUserLoggedIn := role >= 0
	isUserAdmin := userRole > 1

	if !isUserLoggedIn || userNickname == "" {
		utils.Error("App gate: User is not logged in", nil)
		return performLogin(w, req, route)
	}

	if isUserLoggedIn && !isUserAdmin && userNickname != "" {
		utils.Error("App gate:  User is not Authorized (not admin)", nil)
		utils.HTTPError(w, "User not Authorized", http.StatusUnauthorized, "HTTP004")
		return errors.New("User is not Admin")
	}

	return nil
}
