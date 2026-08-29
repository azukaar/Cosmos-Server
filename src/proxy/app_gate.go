package proxy

import (
	"errors"
	"net/http"
	"fmt"
	"net/url"
	"crypto/sha256"
	"encoding/base64"

	"github.com/azukaar/cosmos-server/src/utils"
)

func performLogin(w http.ResponseWriter, req *http.Request, route utils.ProxyRouteConfig) error {
	config := utils.GetMainConfig()
	// keyed on host+path so same-path apps on different hosts don't collide
	pathHash := sha256.Sum256([]byte(route.Host + "\x00" + req.URL.Path + config.HTTPConfig.AuthPrivateKey[32:64]))
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
			
			// The auto-provisioned client is public, so PKCE is mandatory. Derive the
			// verifier deterministically from host+path (recomputed in detectCallbackEndpoint)
			// and send its S256 challenge. base64url output needs no URL-escaping.
			verifier := utils.DerivePKCEVerifier(route.Host, req.URL.Path)
			challengeHash := sha256.Sum256([]byte(verifier))
			codeChallenge := base64.RawURLEncoding.EncodeToString(challengeHash[:])

			//TODO: State should be a random string
			authURL := fmt.Sprintf("%s/cosmos-ui/openid?"+
					"response_type=code&"+
					"client_id=%s&"+
					"redirect_uri=%s&"+
					"scope=openid&"+
					"state=%s&"+
					"code_challenge=%s&"+
					"code_challenge_method=S256",
					mainDomain,
					client.ID,
					url.QueryEscape(client.RedirectURIs[0]),
					url.QueryEscape(state),
					codeChallenge,
			)
			
			http.Redirect(w, req, authURL, http.StatusFound)
			return errors.New("User not logged in, redirecting to OpenID")
	}

	return nil
}

// redirectToMFA sends a half-authenticated session to the MFA page on the main domain
func redirectToMFA(w http.ResponseWriter, req *http.Request, mfaState int) error {
	config := utils.GetMainConfig()
	mainDomain := config.HTTPConfig.Hostname
	scheme := "http://"
	if utils.IsHTTPS {
		scheme = "https://"
	}
	// cross-host apps can't redirect back to themselves; bounce to the UI home
	back := "/cosmos-ui"
	if req.Host == mainDomain {
		back = req.URL.Path
	}
	page := "loginmfa"
	if mfaState == 2 {
		page = "newmfa"
	}
	http.Redirect(w, req, scheme+mainDomain+"/cosmos-ui/"+page+"?invalid=1&redirect="+url.QueryEscape(back), http.StatusTemporaryRedirect)
	return errors.New("User requires MFA")
}

func LoggedInOnlyWithRedirect(w http.ResponseWriter, req *http.Request, route utils.ProxyRouteConfig) error {
	if utils.GetAuthContext(req).Nickname == "" {
		utils.Error("App gate: User is not logged in", nil)
		return performLogin(w, req, route)
	}

	if st := utils.GetAuthContext(req).MFAState; st != 0 {
		utils.Error("App gate: MFA required", nil)
		return redirectToMFA(w, req, st)
	}

	return nil
}

func AdminOnlyWithRedirect(w http.ResponseWriter, req *http.Request, route utils.ProxyRouteConfig) error {
	if utils.GetAuthContext(req).Nickname == "" {
		utils.Error("App gate: User is not logged in", nil)
		return performLogin(w, req, route)
	}

	if !utils.HasPermission(req, utils.PERM_ADMIN_READ) {
		utils.Error("App gate: User is not Authorized (not admin)", nil)
		utils.HTTPError(w, "User not Authorized", http.StatusUnauthorized, "HTTP004")
		return errors.New("User is not Admin")
	}

	if st := utils.GetAuthContext(req).MFAState; st != 0 {
		utils.Error("App gate: MFA required", nil)
		return redirectToMFA(w, req, st)
	}

	return nil
}
