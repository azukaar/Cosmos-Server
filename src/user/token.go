package user

import (
	"net/http"
	"../utils"
	"github.com/golang-jwt/jwt"
	"errors"
	"strings"
	"time"
)

func RefreshUserToken(w http.ResponseWriter, req *http.Request) (utils.User, error) {
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
		utils.Error("UserToken: Token likely falsified", errT)
		logOutUser(w)
		redirectToReLogin(w, req)
		return utils.User{}, errors.New("Token likely falsified")
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
		logOutUser(w)
		redirectToReLogin(w, req)
		return utils.User{}, errors.New("Token not valid")
	}

	nickname := claims["nickname"].(string)
	passwordCycle := int(claims["passwordCycle"].(float64))

	userInBase := utils.User{}

	c := utils.GetCollection(utils.GetRootAppId(), "users")
	
	errDB := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&userInBase)
	
	if errDB != nil {
		utils.Error("UserToken: User not found", errDB)
		logOutUser(w)
		redirectToReLogin(w, req)
		return utils.User{}, errors.New("User not found")
	}

	if userInBase.PasswordCycle != passwordCycle {
		utils.Error("UserToken: Password cycle changed, token is too old", nil)
		logOutUser(w)
		redirectToReLogin(w, req)
		return utils.User{}, errors.New("Password cycle changed, token is too old")
	}

	return userInBase, nil
}

func GetUserR(req *http.Request) (string, string) {
	return req.Header.Get("x-cosmos-user"), req.Header.Get("x-cosmos-role")
}

func logOutUser(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name: "jwttoken",
		Value: "",
		Expires: time.Now().Add(-time.Hour * 24 * 365),
		Path: "/",
		Secure: true,
		HttpOnly: true,
		Domain: utils.GetMainConfig().HTTPConfig.Hostname,
	}

	http.SetCookie(w, &cookie)

	// TODO: Remove all other cookies from apps

	// TODO: logout every other device if asked by increasing passwordcycle
}

func redirectToReLogin(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "/ui/login?invalid=1&redirect=" + req.URL.Path, http.StatusFound)
}

func SendUserToken(w http.ResponseWriter, user utils.User) {
	expiration := time.Now().Add(2 * 24 * time.Hour)

	token := jwt.New(jwt.SigningMethodEdDSA)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = expiration.Unix()
	claims["role"] = user.Role
	claims["nickname"] = user.Nickname
	claims["passwordCycle"] = user.PasswordCycle
	claims["iat"] = time.Now().Unix()
	claims["nbf"] = time.Now().Unix()

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
		Secure: true,
		HttpOnly: true,
		Domain: utils.GetMainConfig().HTTPConfig.Hostname,
	}

	http.SetCookie(w, &cookie)
	// http.SetCookie(w, &cookie2)
}