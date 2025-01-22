package authorizationserver

import (
	"encoding/json"
	"net/http"
	// _url "net/url"

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
	config := utils.GetMainConfig()
	hostname, proxyRoute := utils.FindRouteByReqHost(req.Host)
	
	realHostname := config.HTTPConfig.Hostname
	if utils.IsHTTPS {
			realHostname = "https://" + realHostname
	} else {
			realHostname = "http://" + realHostname
	}

	rw.Header().Del("Content-Type")
	rw.Header().Set("Content-Type", "application/json")

	configuration := &oidcConfiguration{
			Issuer:                                 hostname,
			AuthURL:                                realHostname + "/cosmos-ui/openid",
			TokenURL:                               hostname + "/oauth2/token",
			JWKsURI:                                hostname + "/.well-known/jwks.json",
			RevocationEndpoint:                     hostname + "/oauth2/revoke",
			UserinfoEndpoint:                       hostname + "/oauth2/userinfo",
			SubjectTypes:                           []string{"public", "pairwise"},
			ResponseTypes:                          []string{"code", "code id_token", "id_token", "token id_token", "token", "token id_token code"},
			ClaimsSupported:                        []string{"aud", "email", "email_verified", "exp", "iat", "iss", "locale", "name", "sub"},
			ScopesSupported:                        []string{"openid", "offline", "profile", "email", "address", "phone"},
			TokenEndpointAuthMethodsSupported:      []string{"client_secret_post", "client_secret_basic", "private_key_jwt", "none"},
			GrantTypesSupported:                    []string{"authorization_code", "implicit", "client_credentials", "refresh_token"},
			ResponseModesSupported:                 []string{"query", "fragment"},
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
	}

	configurationMap := map[string]interface{}{}

	// convert struct to map
	configurationBytes, _ := json.Marshal(configuration)

	// convert map to json
	json.Unmarshal(configurationBytes, &configurationMap)

	// If this is a proxy domain, add proxy-specific metadata
	if proxyRoute != nil {
		client := utils.GetProxyOIDCredentials(*proxyRoute, false)

		detectMap := map[string]interface{}{}
		detectMap["enabled"] = true

		detectMap["client_id"] = client.ID
		detectMap["redirect_uri"] = client.RedirectURIs
		detectMap["configuration_endpoint"] = realHostname + "/.well-known/openid-configuration"
		detectMap["auth_endpoint"] = realHostname + "/cosmos-ui/oauth"
		detectMap["suggested_redirect"] = "/oauth/callback"

		detectMap["provider_name"] = "Cosmos Cloud"
		detectMap["provider_logo"] = realHostname + "/logo"

		detectMap["login_label"] = "Login with Cosmos Cloud"
		detectMap["login_logo"] = realHostname + "/logo"
		detectMap["login_bg"] = "#000000"
		detectMap["login_color"] = "#ffffff"

		configurationMap["oid_detect"] = detectMap
	}

	// Regular OIDC config for main domain
	json.NewEncoder(rw).Encode(configurationMap)
}