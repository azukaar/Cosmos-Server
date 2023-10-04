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

func quickLoggout(w http.ResponseWriter, req *http.Request, err error) (utils.User, error) {
	utils.Error("UserToken: Token likely falsified", err)
	logOutUser(w, req)
	redirectToReLogin(w, req)
	return utils.User{}, errors.New("Token likely falsified")
}

func RefreshUserToken(w http.ResponseWriter, req *http.Request) (utils.User, error) {
	config := utils.GetMainConfig()
	
	// if new install
	if config.NewInstall {
		// check route
		if req.URL.Path != "/cosmos/api/status" && req.URL.Path != "/cosmos/api/newInstall" && req.URL.Path != "/cosmos/api/dns" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "NEW_INSTALL",
			})
			return utils.User{}, errors.New("New install")
		} else {
			return utils.User{}, nil
		}
	}

	cookie, err := req.Cookie("jwttoken")

	if err != nil {
		return utils.User{}, nil
	}
	
	tokenString := cookie.Value

	if tokenString == "" {
		return utils.User{}, nil
	}
	
	ed25519Key, errK := jwt.ParseEdPublicKeyFromPEM([]byte(utils.GetPublicAuthKey()))

	if errK != nil {
		utils.Error("UserToken: Cannot read auth public key", errK)
		utils.HTTPError(w, "Authorization Error", http.StatusInternalServerError, "A001")
		return utils.User{}, errors.New("Cannot read auth public key")
	}

	parts := strings.Split(tokenString, ".")
	
	errT := jwt.SigningMethodEdDSA.Verify(strings.Join(parts[0:2], "."), parts[2], ed25519Key)

	if errT != nil {
		if _, e := quickLoggout(w, req, errT); e != nil {
			return utils.User{}, errT
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
		return utils.User{}, errors.New("Token not valid")
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
			return utils.User{}, e
		}
	}
	
	if passwordCycleFloat, ok := claims["passwordCycle"].(float64); ok {
		passwordCycle = int(passwordCycleFloat)
	} else {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return utils.User{}, e
		}
	}
	
	if mfaDone, ok = claims["mfaDone"].(bool); !ok {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return utils.User{}, e
		}
	}

	if forDomain, ok = claims["forDomain"].(string); !ok {
		if _, e := quickLoggout(w, req, nil); e != nil {
			return utils.User{}, e
		}
	}
	
	reqHostname := req.Host
	reqHostNoPort := strings.Split(reqHostname, ":")[0]
	
	if !strings.HasSuffix(reqHostNoPort, forDomain) {
		utils.Error("UserToken: token is not valid for this domain", nil)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return utils.User{}, errors.New("JWT Token not valid for this domain")
	}

	userInBase := utils.User{}

	c, errCo := utils.GetCollection(utils.GetRootAppId(), "users")
		if errCo != nil {
				utils.Error("Database Connect", errCo)
				utils.HTTPError(w, "Database", http.StatusInternalServerError, "DB001")
				return utils.User{}, errCo
		}
	
	errDB := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&userInBase)
	
	if errDB != nil {
		utils.Error("UserToken: User not found", errDB)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return utils.User{}, errors.New("User not found")
	}

	if userInBase.PasswordCycle != passwordCycle {
		utils.Error("UserToken: Password cycle changed, token is too old", nil)
		logOutUser(w, req)
		redirectToReLogin(w, req)
		return utils.User{}, errors.New("Password cycle changed, token is too old")
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

	if time.Now().Unix() - int64(claims["iat"].(float64)) > 3600 {
		SendUserToken(w, req, userInBase, mfaDone)
	}

	return userInBase, nil
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
		Secure: utils.IsHTTPS,
		HttpOnly: true,
	}

	clientCookie := http.Cookie{
		Name: "client-infos",
		Value: "{}",
		Expires: time.Now().Add(-time.Hour * 24 * 365),
		Path: "/",
		Secure: utils.IsHTTPS,
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

func SendUserToken(w http.ResponseWriter, req *http.Request, user utils.User, mfaDone bool) {
	reqHostname := req.Host
	reqHostNoPort := strings.Split(reqHostname, ":")[0]

	expiration := time.Now().Add(3 * 24 * time.Hour)

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = expiration.Unix()
	claims["role"] = user.Role
	claims["nickname"] = user.Nickname
	claims["passwordCycle"] = user.PasswordCycle
	claims["iat"] = time.Now().Unix()
	claims["nbf"] = time.Now().Unix()
	claims["mfaDone"] = mfaDone
	claims["forDomain"] = reqHostNoPort

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
		Secure: utils.IsHTTPS,
		HttpOnly: true,
	}

	clientCookie := http.Cookie{
		Name: "client-infos",
		Value: user.Nickname + "," + strconv.Itoa(int(user.Role)),
		Expires: expiration,
		Path: "/",
		Secure: utils.IsHTTPS,
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