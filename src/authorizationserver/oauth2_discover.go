package authorizationserver

import (
	"encoding/json"
	"net/http"
	"fmt"
	"os"

	"github.com/azukaar/cosmos-server/src/utils"
)

type oidcConfiguration struct {
	Issuer string `json:"issuer"`
	AuthURL string `json:"authorization_endpoint"`
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`
	TokenURL string `json:"token_endpoint"`
	JWKsURI string `json:"jwks_uri"`
	SubjectTypes []string `json:"subject_types_supported"`
	ResponseTypes []string `json:"response_types_supported"`
	ClaimsSupported []string `json:"claims_supported"`
	GrantTypesSupported []string `json:"grant_types_supported"`
	ResponseModesSupported []string `json:"response_modes_supported"`
	UserinfoEndpoint string `json:"userinfo_endpoint"`
	ScopesSupported []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	UserinfoSigningAlgValuesSupported []string `json:"userinfo_signing_alg_values_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	IDTokenSignedResponseAlg []string `json:"id_token_signed_response_alg"`
	UserinfoSignedResponseAlg []string `json:"userinfo_signed_response_alg"`
	RequestParameterSupported bool `json:"request_parameter_supported"`
	RequestURIParameterSupported bool `json:"request_uri_parameter_supported"`
	RequireRequestURIRegistration bool `json:"require_request_uri_registration"`
	ClaimsParameterSupported bool `json:"claims_parameter_supported"`
	RevocationEndpoint string `json:"revocation_endpoint"`
	BackChannelLogoutSupported bool `json:"backchannel_logout_supported"`
	BackChannelLogoutSessionSupported bool `json:"backchannel_logout_session_supported"`
	FrontChannelLogoutSupported bool `json:"frontchannel_logout_supported"`
	FrontChannelLogoutSessionSupported bool `json:"frontchannel_logout_session_supported"`
	EndSessionEndpoint string `json:"end_session_endpoint"`
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}


func discoverEndpoint(rw http.ResponseWriter, req *http.Request) {
	// get hostname from request
	hostname := req.Host
	
	realHostname := utils.GetMainConfig().HTTPConfig.Hostname
	if utils.IsHTTPS {
		realHostname = "https://" + realHostname
	} else {
		realHostname = "http://" + realHostname
	}

	// external request
	if hostname == utils.GetMainConfig().HTTPConfig.Hostname {
		if utils.IsHTTPS {
			hostname = "https://" + hostname
		} else {
			hostname = "http://" + hostname
		}
	} else if hostname == os.Getenv("HOSTNAME") {
		hostname = "http://" + hostname
	} else {
		utils.Error(fmt.Sprintf("invalid hostname for OpenID: %s expecting %s or %s", hostname, utils.GetMainConfig().HTTPConfig.Hostname, os.Getenv("HOSTNAME")), nil)
		http.Error(rw, "invalid hostname for OpenID", http.StatusBadRequest)
		return
	}



	json.NewEncoder(rw).Encode(&oidcConfiguration{
		Issuer:                                 hostname,
		AuthURL:                                realHostname + "/cosmos-ui/openid",
		TokenURL:                               hostname + "/oauth2/token",
		JWKsURI:                                hostname + "/.well-known/jwks.json",
		RevocationEndpoint:                     hostname + "/oauth2/revoke",
		UserinfoEndpoint:                       hostname + "/oauth2/userinfo",
		// RegistrationEndpoint:                   hostname + "/oauth2/register",
		SubjectTypes:                           []string{"public", "pairwise"},
		ResponseTypes:                          []string{"code", "code id_token", "id_token", "token id_token", "token", "token id_token code"},
		ClaimsSupported:                        []string{"aud", "email", "email_verified", "exp", "iat", "iss", "locale", "name", "sub"},
		ScopesSupported:                        []string{"openid", "offline", "profile", "email", "address", "phone"},
		TokenEndpointAuthMethodsSupported:      []string{"client_secret_post", "client_secret_basic", "private_key_jwt", "none"},
		// IDTokenSigningAlgValuesSupported:       []string{key.Algorithm},
		// IDTokenSignedResponseAlg:               []string{key.Algorithm},
		// UserinfoSignedResponseAlg:              []string{key.Algorithm},
		GrantTypesSupported:                    []string{"authorization_code", "implicit", "client_credentials", "refresh_token"},
		ResponseModesSupported:                 []string{"query", "fragment"},
		// UserinfoSigningAlgValuesSupported:      []string{"none", key.Algorithm},
		RequestParameterSupported:              true,
		RequestURIParameterSupported:           true,
		RequireRequestURIRegistration:          true,
		BackChannelLogoutSupported:             true,
		BackChannelLogoutSessionSupported:      true,
		FrontChannelLogoutSupported:            true,
		FrontChannelLogoutSessionSupported:     true,
		EndSessionEndpoint:                     hostname + "/cosmos-ui/logout",
		RequestObjectSigningAlgValuesSupported: []string{"RS256"},
		CodeChallengeMethodsSupported:          []string{"plain", "S256"},
	})
}
