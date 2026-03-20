package user

import (
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/golang-jwt/jwt"
	"errors"
	"strconv"
	"strings"
	"time"
	"encoding/json"
)

func shouldCookieBeSecured(ip string) bool {
	config := utils.GetMainConfig()

	if !utils.IsHTTPS {
		return false
	} else {
		if config.HTTPConfig.AllowHTTPLocalIPAccess && utils.IsLocalIP(ip) {
			return false
		} else {
			return true
		}
	}
}

func quickLoggout(w http.ResponseWriter, req *http.Request, err error) (utils.User, error) {
	utils.Error("UserToken: Token likely falsified", err)
	logOutUser(w, req)
	redirectToReLogin(w, req)
	return utils.User{}, errors.New("Token likely falsified")
}

func RefreshUserToken(w http.ResponseWriter, req *http.Request) ([]utils.Permission, bool, utils.User, error) {
	config := utils.GetMainConfig()
	
	// if new install
	if config.NewInstall {
		// check route
		if req.URL.Path != "/cosmos/api/status" && req.URL.Path != "/cosmos/api/newInstall" && req.URL.Path != "/cosmos/api/dns" && req.URL.Path != "/cosmos/api/setup" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "NEW_INSTALL",
			})
			return nil, false, utils.User{}, errors.New("New install")
		} else {
			return nil, false, utils.User{}, nil
		}
	}

	cookie, err := req.Cookie("jwttoken")

	if err != nil {
		return nil, false, utils.User{}, nil
	}

	tokenString := cookie.Value

	if tokenString == "" {
		return nil, false, utils.User{}, nil
	}

	ed25519Key, errK := jwt.ParseEdPublicKeyFromPEM([]byte(utils.GetPublicAuthKey()))

	if errK != nil {
		utils.Error("UserToken: Cannot read auth public key", errK)
		utils.HTTPError(w, "Authorization Error", http.StatusInternalServerError, "A001")
		return nil, false, utils.User{}, errors.New("Cannot read auth public key")
	}

	parts := strings.Split(tokenString, ".")

	errT := jwt.SigningMethodEdDSA.Verify(strings.Join(parts[0:2], "."), parts[2], ed25519Key)

	if errT != nil {
		if _, e := quickLoggout(w, req, errT); e != nil {
			return nil, false, utils.User{}, errT
		}
	}

	type claimsType struct {
		nickname string
	}

	claims := jwt.MapClaims{}

	_, errP := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
    return ed25519Key, nil
	})

	if errP != nil {
		utils.Error("UserToken: token is not valid", nil)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return nil, false, utils.User{}, errors.New("Token not valid")
	}

	var (
		nickname      string
		passwordCycle int
		mfaDone       bool
		ok            bool
		forDomain     string
	)

	if nickname, ok = claims["nickname"].(string); !ok {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return nil, false, utils.User{}, e
		}
	}

	if passwordCycleFloat, ok := claims["passwordCycle"].(float64); ok {
		passwordCycle = int(passwordCycleFloat)
	} else {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return nil, false, utils.User{}, e
		}
	}

	if mfaDone, ok = claims["mfaDone"].(bool); !ok {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return nil, false, utils.User{}, e
		}
	}

	if forDomain, ok = claims["forDomain"].(string); !ok {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return nil, false, utils.User{}, e
		}
	}

	reqHostname := req.Host
	reqHostNoPort := strings.Split(reqHostname, ":")[0]

	if !strings.HasSuffix(reqHostNoPort, forDomain) {
		utils.Error("UserToken: token is not valid for this domain", nil)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return nil, false, utils.User{}, errors.New("JWT Token not valid for this domain")
	}

	userInBase := utils.User{}

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  defer closeDb()

	if errCo != nil {
			utils.Error("Database Connect", errCo)
			utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
			return nil, false, utils.User{}, errCo
	}

	errDB := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&userInBase)

	if errDB != nil {
		utils.Error("UserToken: User not found", errDB)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return nil, false, utils.User{}, errors.New("User not found")
	}

	if userInBase.PasswordCycle != passwordCycle {
		utils.Error("UserToken: Password cycle changed, token is too old", nil)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return nil, false, utils.User{}, errors.New("Password cycle changed, token is too old")
	}

	requestURL := req.URL.Path
	isSettingMFA := strings.HasPrefix(requestURL, "/cosmos-ui/loginmfa") || strings.HasPrefix(requestURL, "/cosmos-ui/newmfa") || strings.HasPrefix(requestURL, "/api/mfa")

	userInBase.MFAState = 0

	if !isSettingMFA && (userInBase.MFAKey != "" && userInBase.Was2FAVerified && !mfaDone) {
		utils.Warn("UserToken: MFA required")
		userInBase.MFAState = 1
	} else if !isSettingMFA && (config.RequireMFA && !mfaDone) {
		utils.Warn("UserToken: MFA not set")
		userInBase.MFAState = 2
	}

	// Parse permissions from JWT token
	rawPerms, hasPerms := claims["permissions"].([]interface{})
	if !hasPerms {
		// Old token without permissions claim — force re-login
		utils.Warn("UserToken: Token missing permissions claim, forcing re-login")
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return nil, false, utils.User{}, errors.New("Token missing permissions claim")
	}
	permissions := make([]utils.Permission, len(rawPerms))
	for i, rp := range rawPerms {
		permissions[i] = utils.Permission(int(rp.(float64)))
	}

	// Determine sudo state from sudo-until timestamp
	isSudoed := false
	hasSudoPerms := utils.PermissionsHaveSudo(permissions)

	if hasSudoPerms {
		if sudoUntilRaw, ok := claims["sudo-until"]; ok {
			sudoUntil := int64(sudoUntilRaw.(float64))
			if sudoUntil > time.Now().Unix() {
				isSudoed = true
				// If close to sudo expiry (within 1 hour), refresh
				if time.Now().Unix() + 3600 > sudoUntil {
					SendUserToken(w, req, userInBase, mfaDone, true)
					utils.Debug("UserToken: Sudo refreshing")
				}
			}
			// else: sudo expired, isSudoed stays false
		}
		// else: no sudo-until claim, isSudoed stays false
	}

	// if close to expiration, refresh
	if int64(claims["iat"].(float64)) + (24 * 3600) < time.Now().Unix() {
		SendUserToken(w, req, userInBase, mfaDone, isSudoed)
	}

	return permissions, isSudoed, userInBase, nil
}


func logOutUser(w http.ResponseWriter, req *http.Request) {
	reqHostname := req.Host
	reqHostNoPort := strings.Split(reqHostname, ":")[0]

	cookie := http.Cookie{
		Name: "jwttoken",
		Value: "",
		Expires: time.Now().Add(-time.Hour * 24 * 365),
		Path: "/",
		Secure: shouldCookieBeSecured(req.RemoteAddr),
		HttpOnly: true,
	}

	clientCookie := http.Cookie{
		Name: "client-infos",
		Value: "{}",
		Expires: time.Now().Add(-time.Hour * 24 * 365),
		Path: "/",
		Secure: shouldCookieBeSecured(req.RemoteAddr),
		HttpOnly: false,
	}

	if reqHostNoPort == "localhost" || reqHostNoPort == "0.0.0.0" {
		cookie.Domain = ""
		clientCookie.Domain = ""
	} else {
		cookie.Domain = "." + reqHostNoPort
		clientCookie.Domain = "." + reqHostNoPort
	}

	http.SetCookie(w, &cookie)
	http.SetCookie(w, &clientCookie)

	// TODO: logout every other device if asked by increasing passwordcycle
}

func redirectToReLogin(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "/cosmos-ui/login?invalid=1&redirect=" + req.URL.Path + "&" + req.URL.RawQuery, http.StatusTemporaryRedirect)
}

func redirectToLoginMFA(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "/cosmos-ui/loginmfa?invalid=1&redirect=" + req.URL.Path + "&" + req.URL.RawQuery, http.StatusTemporaryRedirect)
}

func redirectToNewMFA(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "/cosmos-ui/newmfa?invalid=1&redirect=" + req.URL.Path + "&" + req.URL.RawQuery, http.StatusTemporaryRedirect)
}

func SendUserToken(w http.ResponseWriter, req *http.Request, user utils.User, mfaDone bool, sudoed bool) {
	reqHostname := req.Host
	reqHostNoPort := strings.Split(reqHostname, ":")[0]

	expiration := time.Now().Add(14 * 24 * time.Hour)

	// Resolve permissions from role at token creation time
	perms := utils.GetRolePermissions(user.Role)
	permInts := make([]interface{}, len(perms))
	for i, p := range perms {
		permInts[i] = int(p)
	}

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = expiration.Unix()
	claims["permissions"] = permInts
	claims["nickname"] = user.Nickname
	claims["passwordCycle"] = user.PasswordCycle
	claims["iat"] = time.Now().Unix()
	claims["nbf"] = time.Now().Unix()
	claims["mfaDone"] = mfaDone
	claims["forDomain"] = reqHostNoPort

	sudoUntil := int64(-1)

	// if sudoed, add a timeout
	if sudoed {
		sudoUntil = time.Now().Add(time.Hour * 2).Unix()
		claims["sudo-until"] = sudoUntil
	}

	key, err5 := jwt.ParseEdPrivateKeyFromPEM([]byte(utils.GetPrivateAuthKey()))
	
	if err5 != nil {
		utils.Error("UserLogin: Error while retrieving signing key", err5)
		utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
		return
	}

	tokenString, err4 := token.SignedString(key)

	if err4 != nil {
		utils.Error("UserLogin: Error while signing token", err4)
		utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
		return
	}

	cookie := http.Cookie{
		Name: "jwttoken",
		Value: tokenString,
		Expires: expiration,
		Path: "/",
		Secure: shouldCookieBeSecured(req.RemoteAddr),
		HttpOnly: true,
	}

	// Build permissions string for client cookie (reuse perms from above)
	permStrs := make([]string, len(perms))
	for i, p := range perms {
		permStrs[i] = strconv.Itoa(int(p))
	}
	permsString := strings.Join(permStrs, "-")

	clientCookie := http.Cookie{
		Name: "client-infos",
		Value: user.Nickname + "," + permsString + "," + strconv.Itoa(int(sudoUntil)),
		Expires: expiration,
		Path: "/",
		Secure: shouldCookieBeSecured(req.RemoteAddr),
		HttpOnly: false,
	}

	utils.Log("UserLogin: Setting cookie for " + reqHostNoPort)

	if reqHostNoPort == "localhost" || reqHostNoPort == "0.0.0.0" {
		cookie.Domain = ""
		clientCookie.Domain = ""
	} else {
		if utils.IsValidHostname(reqHostNoPort) {
			cookie.Domain = "." + reqHostNoPort
			clientCookie.Domain = "." + reqHostNoPort
		} else {
			utils.Error("UserLogin: Invalid hostname", nil)
			utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
			return
		}
	}

	http.SetCookie(w, &cookie)
	http.SetCookie(w, &clientCookie)
}