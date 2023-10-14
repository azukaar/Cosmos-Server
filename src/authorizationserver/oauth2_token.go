package authorizationserver

import (
	"net/http"
	// "fmt"

	"github.com/azukaar/cosmos-server/src/utils"
)

func tokenEndpoint(rw http.ResponseWriter, req *http.Request) {
	utils.Log("Token endpoint")

	// This context will be passed to all methods.
	ctx := req.Context()

	// Create an empty session object which will be passed to the request handlers
	mySessionData := newSession("", req)

	// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
	accessRequest, err := oauth2.NewAccessRequest(ctx, req, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		utils.Error("Error occurred in NewAccessRequest", err)
		oauth2.WriteAccessError(ctx, rw, accessRequest, err)
		return
	}

	// If this is a client_credentials grant, grant all requested scopes
	// NewAccessRequest validated that all requested scopes the client is allowed to perform
	// based on configured scope matching strategy.
	if accessRequest.GetGrantTypes().ExactOne("client_credentials") {
		for _, scope := range accessRequest.GetRequestedScopes() {
			accessRequest.GrantScope(scope)
		}
	}

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		utils.Error("Error occurred in NewAccessResponse", err)
		oauth2.WriteAccessError(ctx, rw, accessRequest, err)
		return
	}

	utils.Log("Access token granted to client: " + accessRequest.GetClient().GetID())

	// All done, send the response.
	oauth2.WriteAccessResponse(ctx, rw, accessRequest, response)

	// The client now has a valid access token
}
