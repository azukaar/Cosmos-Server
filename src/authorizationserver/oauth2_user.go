package authorizationserver

import (
	"encoding/json"
	"net/http"
	"github.com/ory/fosite"
	"fmt"

	"github.com/ory/fosite/handler/openid"

	"github.com/azukaar/cosmos-server/src/utils"

)

// getUserClaims builds the OIDC profile claims for a user. It is the single
// source of truth shared by the userinfo endpoint and the id_token, so both
// always return the same set of claims.
func getUserClaims(nickname string, scopes fosite.Arguments) (map[string]interface{}, error) {
	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
	defer closeDb()
	if errCo != nil {
		return nil, errCo
	}

	user := utils.User{}
	err := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&user)
	if err != nil {
		return nil, err
	}

	claims := map[string]interface{}{
		"name":               nickname,
		"username":           nickname,
		"nickname":           nickname,
		"preferred_username": nickname,
	}

	if scopes.Has("email") {
		claims["email"] = user.Email
	}

	// Derive the role from the fetched user record rather than the request's
	// auth context: the userinfo endpoint only carries a Bearer token (no Cosmos
	// session), so HasPermission(req, ...) would always be false there.
	if user.Role >= utils.ADMIN {
		claims["role"] = "admin"
	} else {
		claims["role"] = "user"
	}

	return claims, nil
}

func userInfosEndpoint(rw http.ResponseWriter, req *http.Request) {	
	ctx := req.Context()
	mySessionData := newSession("", req)
	tokenType, ar, err  := oauth2.IntrospectToken(ctx, fosite.AccessTokenFromRequest(req), fosite.AccessToken, mySessionData)

	if err != nil {
		// log.Printf("Error occurred in NewIntrospectionRequest: %+v", err)
		oauth2.WriteIntrospectionError(ctx, rw, err)
		return
	}
	
	if tokenType != fosite.AccessToken {
		errorDescription := "Only access tokens are allowed in the authorization header."
		rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer error="invalid_token",error_description="%s"`, errorDescription))
		// h.r.Writer().WriteErrorCode(w, r, http.StatusUnauthorized, errors.New(errorDescription))
		utils.Error("UserInfosGet: Only access tokens are allowed in the authorization header", err)
		utils.HTTPError(rw, "Only access tokens are allowed in the authorization header", http.StatusInternalServerError, "UD001")
		return
	}

	interim := ar.GetSession().(*openid.DefaultSession).IDTokenClaims().ToMap()

	nickname := interim["sub"].(string)

	utils.Debug("UserInfosGet: Get user " + nickname)

	claims, err := getUserClaims(nickname, ar.GetGrantedScopes())
	if err != nil {
		utils.Error("UserInfosGet: Error while getting user", err)
		utils.HTTPError(rw, "User Get Error", http.StatusInternalServerError, "UD001")
		return
	}

	claims["sub"] = nickname
	claims["iat"] = interim["iat"]
	claims["exp"] = interim["exp"]
	claims["iss"] = interim["iss"]

	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(rw).Encode(claims)
}
