package authorizationserver

import (
	"encoding/json"
	"net/http"
	"fmt"
	"os"

	"github.com/azukaar/cosmos-server/src/utils"
)


// OpenID Connect Discovery Metadata
//
// Includes links to several endpoints (for example `/oauth2/token`) and exposes information on supported signature algorithms
// among others.
//
// swagger:model oidcConfiguration
type oidcConfiguration struct {
	// OpenID Connect Issuer URL
	//
	// An URL using the https scheme with no query or fragment component that the OP asserts as its IssuerURL Identifier.
	// If IssuerURL discovery is supported , this value MUST be identical to the issuer value returned
	// by WebFinger. This also MUST be identical to the iss Claim value in ID Tokens issued from this IssuerURL.
	//
	// required: true
	// example: https://playground.ory.sh/ory-hydra/public/
	Issuer string `json:"issuer"`

	// OAuth 2.0 Authorization Endpoint URL
	//
	// required: true
	// example: https://playground.ory.sh/ory-hydra/public/oauth2/auth
	AuthURL string `json:"authorization_endpoint"`

	// OpenID Connect Dynamic Client Registration Endpoint URL
	//
	// example: https://playground.ory.sh/ory-hydra/admin/client
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`

	// OAuth 2.0 Token Endpoint URL
	//
	// required: true
	// example: https://playground.ory.sh/ory-hydra/public/oauth2/token
	TokenURL string `json:"token_endpoint"`

	// OpenID Connect Well-Known JSON Web Keys URL
	//
	// URL of the OP's JSON Web Key Set [JWK] document. This contains the signing key(s) the RP uses to validate
	// signatures from the OP. The JWK Set MAY also contain the Server's encryption key(s), which are used by RPs
	// to encrypt requests to the Server. When both signing and encryption keys are made available, a use (Key Use)
	// parameter value is REQUIRED for all keys in the referenced JWK Set to indicate each key's intended usage.
	// Although some algorithms allow the same key to be used for both signatures and encryption, doing so is
	// NOT RECOMMENDED, as it is less secure. The JWK x5c parameter MAY be used to provide X.509 representations of
	// keys provided. When used, the bare key values MUST still be present and MUST match those in the certificate.
	//
	// required: true
	// example: https://{slug}.projects.oryapis.com/.well-known/jwks.json
	JWKsURI string `json:"jwks_uri"`

	// OpenID Connect Supported Subject Types
	//
	// JSON array containing a list of the Subject Identifier types that this OP supports. Valid types include
	// pairwise and public.
	//
	// required: true
	// example:
	//   - public
	//   - pairwise
	SubjectTypes []string `json:"subject_types_supported"`

	// OAuth 2.0 Supported Response Types
	//
	// JSON array containing a list of the OAuth 2.0 response_type values that this OP supports. Dynamic OpenID
	// Providers MUST support the code, id_token, and the token id_token Response Type values.
	//
	// required: true
	ResponseTypes []string `json:"response_types_supported"`

	// OpenID Connect Supported Claims
	//
	// JSON array containing a list of the Claim Names of the Claims that the OpenID Provider MAY be able to supply
	// values for. Note that for privacy or other reasons, this might not be an exhaustive list.
	ClaimsSupported []string `json:"claims_supported"`

	// OAuth 2.0 Supported Grant Types
	//
	// JSON array containing a list of the OAuth 2.0 Grant Type values that this OP supports.
	GrantTypesSupported []string `json:"grant_types_supported"`

	// OAuth 2.0 Supported Response Modes
	//
	// JSON array containing a list of the OAuth 2.0 response_mode values that this OP supports.
	ResponseModesSupported []string `json:"response_modes_supported"`

	// OpenID Connect Userinfo URL
	//
	// URL of the OP's UserInfo Endpoint.
	UserinfoEndpoint string `json:"userinfo_endpoint"`

	// OAuth 2.0 Supported Scope Values
	//
	// JSON array containing a list of the OAuth 2.0 [RFC6749] scope values that this server supports. The server MUST
	// support the openid scope value. Servers MAY choose not to advertise some supported scope values even when this parameter is used
	ScopesSupported []string `json:"scopes_supported"`

	// OAuth 2.0 Supported Client Authentication Methods
	//
	// JSON array containing a list of Client Authentication methods supported by this Token Endpoint. The options are
	// client_secret_post, client_secret_basic, client_secret_jwt, and private_key_jwt, as described in Section 9 of OpenID Connect Core 1.0
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`

	// OpenID Connect Supported Userinfo Signing Algorithm
	//
	// 	JSON array containing a list of the JWS [JWS] signing algorithms (alg values) [JWA] supported by the UserInfo Endpoint to encode the Claims in a JWT [JWT].
	UserinfoSigningAlgValuesSupported []string `json:"userinfo_signing_alg_values_supported"`

	// OpenID Connect Supported ID Token Signing Algorithms
	//
	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for the ID Token
	// to encode the Claims in a JWT.
	//
	// required: true
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`

	// OpenID Connect Default ID Token Signing Algorithms
	//
	// Algorithm used to sign OpenID Connect ID Tokens.
	//
	// required: true
	IDTokenSignedResponseAlg []string `json:"id_token_signed_response_alg"`

	// OpenID Connect User Userinfo Signing Algorithm
	//
	// Algorithm used to sign OpenID Connect Userinfo Responses.
	//
	// required: true
	UserinfoSignedResponseAlg []string `json:"userinfo_signed_response_alg"`

	// OpenID Connect Request Parameter Supported
	//
	// Boolean value specifying whether the OP supports use of the request parameter, with true indicating support.
	RequestParameterSupported bool `json:"request_parameter_supported"`

	// OpenID Connect Request URI Parameter Supported
	//
	// Boolean value specifying whether the OP supports use of the request_uri parameter, with true indicating support.
	RequestURIParameterSupported bool `json:"request_uri_parameter_supported"`

	// OpenID Connect Requires Request URI Registration
	//
	// Boolean value specifying whether the OP requires any request_uri values used to be pre-registered
	// using the request_uris registration parameter.
	RequireRequestURIRegistration bool `json:"require_request_uri_registration"`

	// OpenID Connect Claims Parameter Parameter Supported
	//
	// Boolean value specifying whether the OP supports use of the claims parameter, with true indicating support.
	ClaimsParameterSupported bool `json:"claims_parameter_supported"`

	// OAuth 2.0 Token Revocation URL
	//
	// URL of the authorization server's OAuth 2.0 revocation endpoint.
	RevocationEndpoint string `json:"revocation_endpoint"`

	// OpenID Connect Back-Channel Logout Supported
	//
	// Boolean value specifying whether the OP supports back-channel logout, with true indicating support.
	BackChannelLogoutSupported bool `json:"backchannel_logout_supported"`

	// OpenID Connect Back-Channel Logout Session Required
	//
	// Boolean value specifying whether the OP can pass a sid (session ID) Claim in the Logout Token to identify the RP
	// session with the OP. If supported, the sid Claim is also included in ID Tokens issued by the OP
	BackChannelLogoutSessionSupported bool `json:"backchannel_logout_session_supported"`

	// OpenID Connect Front-Channel Logout Supported
	//
	// Boolean value specifying whether the OP supports HTTP-based logout, with true indicating support.
	FrontChannelLogoutSupported bool `json:"frontchannel_logout_supported"`

	// OpenID Connect Front-Channel Logout Session Required
	//
	// Boolean value specifying whether the OP can pass iss (issuer) and sid (session ID) query parameters to identify
	// the RP session with the OP when the frontchannel_logout_uri is used. If supported, the sid Claim is also
	// included in ID Tokens issued by the OP.
	FrontChannelLogoutSessionSupported bool `json:"frontchannel_logout_session_supported"`

	// OpenID Connect End-Session Endpoint
	//
	// URL at the OP to which an RP can perform a redirect to request that the End-User be logged out at the OP.
	EndSessionEndpoint string `json:"end_session_endpoint"`

	// OpenID Connect Supported Request Object Signing Algorithms
	//
	// JSON array containing a list of the JWS signing algorithms (alg values) supported by the OP for Request Objects,
	// which are described in Section 6.1 of OpenID Connect Core 1.0 [OpenID.Core]. These algorithms are used both when
	// the Request Object is passed by value (using the request parameter) and when it is passed by reference
	// (using the request_uri parameter).
	RequestObjectSigningAlgValuesSupported []string `json:"request_object_signing_alg_values_supported"`

	// OAuth 2.0 PKCE Supported Code Challenge Methods
	//
	// JSON array containing a list of Proof Key for Code Exchange (PKCE) [RFC7636] code challenge methods supported
	// by this authorization server.
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
		AuthURL:                                realHostname + "/ui/openid",
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
		EndSessionEndpoint:                     hostname + "/ui/logout",
		RequestObjectSigningAlgValuesSupported: []string{"RS256"},
		CodeChallengeMethodsSupported:          []string{"plain", "S256"},
	})
}
