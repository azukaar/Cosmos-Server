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

func RefreshUserToken(w http.ResponseWriter, req *http.Request) (utils.Role, utils.User, error) {
	config := utils.GetMainConfig()
	
	// if new install
	if config.NewInstall {
		// check route
		if req.URL.Path != "/cosmos/api/status" && req.URL.Path != "/cosmos/api/newInstall" && req.URL.Path != "/cosmos/api/dns" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "NEW_INSTALL",
			})
			return utils.NOONE, utils.User{}, errors.New("New install")
		} else {
			return utils.NOONE, utils.User{}, nil
		}
	}

	cookie, err := req.Cookie("jwttoken")

	if err != nil {
		return utils.NOONE, utils.User{}, nil
	}
	
	tokenString := cookie.Value

	if tokenString == "" {
		return utils.NOONE, utils.User{}, nil
	}
	
	ed25519Key, errK := jwt.ParseEdPublicKeyFromPEM([]byte(utils.GetPublicAuthKey()))

	if errK != nil {
		utils.Error("UserToken: Cannot read auth public key", errK)
		utils.HTTPError(w, "Authorization Error", http.StatusInternalServerError, "A001")
		return utils.NOONE, utils.User{}, errors.New("Cannot read auth public key")
	}

	parts := strings.Split(tokenString, ".")
	
	errT := jwt.SigningMethodEdDSA.Verify(strings.Join(parts[0:2], "."), parts[2], ed25519Key)

	if errT != nil {
		if _, e := quickLoggout(w, req, errT); e != nil {
			return utils.NOONE, utils.User{}, errT
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
		return utils.NOONE, utils.User{}, errors.New("Token not valid")
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
			return utils.NOONE, utils.User{}, e
		}
	}
	
	if passwordCycleFloat, ok := claims["passwordCycle"].(float64); ok {
		passwordCycle = int(passwordCycleFloat)
	} else {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return utils.NOONE, utils.User{}, e
		}
	}
	
	if mfaDone, ok = claims["mfaDone"].(bool); !ok {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return utils.NOONE, utils.User{}, e
		}
	}

	if forDomain, ok = claims["forDomain"].(string); !ok {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return utils.NOONE, utils.User{}, e
		}
	}
	
	reqHostname := req.Host
	reqHostNoPort := strings.Split(reqHostname, ":")[0]
	
	if !strings.HasSuffix(reqHostNoPort, forDomain) {
		utils.Error("UserToken: token is not valid for this domain", nil)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return utils.NOONE, utils.User{}, errors.New("JWT Token not valid for this domain")
	}

	userInBase := utils.User{}

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  defer closeDb()
	
	if errCo != nil {
			utils.Error("Database Connect", errCo)
			utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
			return utils.NOONE, utils.User{}, errCo
	}

	errDB := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&userInBase)
	
	if errDB != nil {
		utils.Error("UserToken: User not found", errDB)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return utils.NOONE, utils.User{}, errors.New("User not found")
	}

	if userInBase.PasswordCycle != passwordCycle {
		utils.Error("UserToken: Password cycle changed, token is too old", nil)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return utils.NOONE, utils.User{}, errors.New("Password cycle changed, token is too old")
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

	tokenRole := (utils.Role)(claims["role"].(float64))

	// if sudo-until exists 
	if (utils.Role)(claims["role"].(float64)) == utils.ADMIN {
		if _, ok := claims["sudo-until"]; ok {
			// if expires, refresh with demotion
			if int64(claims["sudo-until"].(float64)) < time.Now().Unix() {
				tokenRole = utils.USER
				SendUserToken(w, req, userInBase, mfaDone, utils.USER)
				utils.Debug("UserToken: Sudo expired")
			} else if time.Now().Unix() + 3600 > int64(claims["sudo-until"].(float64)) {
				SendUserToken(w, req, userInBase, mfaDone, (utils.Role)(claims["role"].(float64)))
				utils.Debug("UserToken: Sudo refreshing")
			}
		} else {
			tokenRole = utils.USER
			SendUserToken(w, req, userInBase, mfaDone, utils.USER)
			utils.Debug("UserToken: Sudo expired")
		}
	}

	// if close to expiration, refresh
	if int64(claims["iat"].(float64)) + (24 * 3600) < time.Now().Unix() {
		SendUserToken(w, req, userInBase, mfaDone, tokenRole)
	}

	return tokenRole, userInBase, nil
}

func GetUserR(req *http.Request) (string, string) {
	return req.Header.Get("x-cosmos-user"), req.Header.Get("x-cosmos-role")
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

func SendUserToken(w http.ResponseWriter, req *http.Request, user utils.User, mfaDone bool, tokenRole utils.Role) {
	reqHostname := req.Host
	reqHostNoPort := strings.Split(reqHostname, ":")[0]

	expiration := time.Now().Add(14 * 24 * time.Hour)

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = expiration.Unix()
	claims["role"] = tokenRole
	claims["user-role"] = user.Role
	claims["nickname"] = user.Nickname
	claims["passwordCycle"] = user.PasswordCycle
	claims["iat"] = time.Now().Unix()
	claims["nbf"] = time.Now().Unix()
	claims["mfaDone"] = mfaDone
	claims["forDomain"] = reqHostNoPort

	sudoUntil :=  time.Now().Add(time.Second * 5).Unix()
	
	// if role is ADMIN, add a timeout
	if tokenRole == utils.ADMIN {
		claims["sudo-until"] = sudoUntil
	} else {
		sudoUntil = -1;
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

	clientCookie := http.Cookie{
		Name: "client-infos",
		Value: user.Nickname + "," + strconv.Itoa(int(user.Role)) + "," + strconv.Itoa(int(tokenRole)) + "," + strconv.Itoa(int(sudoUntil)),
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