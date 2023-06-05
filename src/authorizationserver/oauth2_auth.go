package authorizationserver

import (
	"net/http"
	"github.com/azukaar/cosmos-server/src/utils"
)

func authEndpoint(rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := req.Context()
	
	if utils.LoggedInOnly(rw, req) != nil {
		return
	}

	nickname := req.Header.Get("x-cosmos-user")
	
	hostname := utils.GetMainConfig().HTTPConfig.Hostname
	if utils.IsHTTPS {
		hostname = "https://" + hostname
	} else {
		hostname = "http://" + hostname
	}

	// Let's create an AuthorizeRequest object!
	// It will analyze the request and extract important information like scopes, response type and others.
	ar, err := oauth2.NewAuthorizeRequest(ctx, req)
	if err != nil {
		utils.Error("Error occurred in NewAuthorizeRequest:", err)
		oauth2.WriteAuthorizeError(ctx, rw, ar, err)
		return
	}

	// let's see what scopes the user gave consent to
	for _, scope := range req.PostForm["scopes"] {
		ar.GrantScope(scope)
	}

	// Now that the user is authorized, we set up a session:
	mySessionData := newSession(nickname, req)

	// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
	// NewAuthorizeResponse is capable of running multiple response type handlers which in turn enables this library
	// to support open id connect.
	response, err := oauth2.NewAuthorizeResponse(ctx, ar, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	if err != nil {
		utils.Error("Error occurred in NewAuthorizeResponse:", err)
		oauth2.WriteAuthorizeError(ctx, rw, ar, err)
		return
	}

	// Last but not least, send the response!
	oauth2.WriteAuthorizeResponse(ctx, rw, ar, response)
}
