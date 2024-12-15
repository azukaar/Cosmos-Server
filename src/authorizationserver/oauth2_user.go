package authorizationserver

import (
	"encoding/json"
	"net/http"
	"github.com/ory/fosite"
	"fmt"

	"github.com/ory/fosite/handler/openid"

	"github.com/azukaar/cosmos-server/src/utils"

)

type oidcUser struct {
	Name string `json:"name"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role string `json:"role"`
	Email string `json:"email"`
	Subject string `json:"sub"`
	IssuedAt int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
	Issuer string `json:"iss"`
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

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
  defer closeDb()
	if errCo != nil {
			utils.Error("Database Connect", errCo)
			utils.HTTPError(rw, "Database", http.StatusInternalServerError, "DB001")
			return
	}

	utils.Debug("UserInfosGet: Get user " + nickname)

	user := utils.User{}

	err = c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&user)

	if err != nil {
		utils.Error("UserInfosGet: Error while getting user", err)
		utils.HTTPError(rw, "User Get Error", http.StatusInternalServerError, "UD001")
		return
	}

	baseToken := &oidcUser{
		Name: interim["sub"].(string),
		Username: interim["sub"].(string),
		Nickname: interim["sub"].(string),
		Subject: interim["sub"].(string),
		IssuedAt: interim["iat"].(int64),
		ExpiresAt: interim["exp"].(int64),
		Issuer: interim["iss"].(string),
	}

	// check scopes has email
	if ar.GetGrantedScopes().Has("email") {
		baseToken.Email = user.Email
	}

	if user.Role == utils.ADMIN {
		baseToken.Role = "admin"
	} else {
		baseToken.Role = "user"
	}
	
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(rw).Encode(baseToken)
}
