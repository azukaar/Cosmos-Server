package authorizationserver

import (
	"math/rand"
	"crypto/rsa"
	"github.com/ory/fosite"
	"time"
	"net/http"
	"os"
	"crypto/sha256"
	"encoding/binary"


	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"

	"github.com/azukaar/cosmos-server/src/utils"

	"github.com/gorilla/mux"
)

var oauth2 fosite.OAuth2Provider
var AuthPrivateKey *rsa.PrivateKey

func Init() {
	config := utils.ReadConfigFromFile()
	authKey := config.HTTPConfig.AuthPrivateKey

	secret := []byte("some-super-cool-secret-that-nobody-knows")
	
	if len(authKey) > 64 {
		secret = []byte(authKey[32:64])
	}
	
	foconfig := &fosite.Config{
		AccessTokenLifespan: time.Minute * 30,
		GlobalSecret:        secret,
	}

	store := storage.NewMemoryStore()
	
	// loop config.OpenIDClients
	for _, client := range config.OpenIDClients {
		utils.Log("Registering OpenID client: " + client.ID)

		// register client
		store.Clients[client.ID] = &fosite.DefaultClient{
			ID:             client.ID,
			Secret:         []byte(client.Secret),
			RedirectURIs:   []string{client.Redirect},
			Scopes:         []string{"openid", "email", "profile", "offline", "roles", "groups", "address", "phone", "role"},
			ResponseTypes:  []string{"id_token", "code", "token", "id_token token", "code id_token", "code token", "code id_token token"},
			GrantTypes:     []string{"implicit", "refresh_token", "authorization_code", "password", "client_credentials"},
		}
	}
	
	// Compute a hash of your authKey to use as a seed for deterministic RSA key generation
	seedHash := sha256.Sum256([]byte(authKey))
	seed := int64(binary.BigEndian.Uint64(seedHash[:]))

	// Now you can create a deterministic source of randomness
	deterministicReader := rand.New(rand.NewSource(seed))

	// privateKey is used to sign JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	AuthPrivateKey, _ = rsa.GenerateKey(deterministicReader, 2048)

	// Build a fosite instance with all OAuth2 and OpenID Connect handlers enabled, plugging in our configurations as specified above.
	oauth2 = compose.ComposeAllEnabled(foconfig, store, AuthPrivateKey)

	utils.Log("OpenID server initialized")
}

func RegisterHandlers(wellKnown *mux.Router, userRouter *mux.Router, serverRouter *mux.Router) {
	// Set up oauth2 endpoints. You could also use gorilla/mux or any other router.
	userRouter.HandleFunc("/auth", authEndpoint)
	serverRouter.HandleFunc("/token", tokenEndpoint)

	// user infos
	serverRouter.HandleFunc("/userinfo", userInfosEndpoint)

	// revoke tokens
	serverRouter.HandleFunc("/revoke", revokeEndpoint)
	serverRouter.HandleFunc("/introspect", introspectionEndpoint)

	// public endpoints
	wellKnown.HandleFunc("/openid-configuration", discoverEndpoint)
	wellKnown.HandleFunc("/jwks.json", jwksEndpoint)
}

// A session is passed from the `/auth` to the `/token` endpoint. You probably want to store data like: "Who made the request",
// "What organization does that person belong to" and so on.
// For our use case, the session will meet the requirements imposed by JWT access tokens, HMAC access tokens and OpenID Connect
// ID Tokens plus a custom field

// newSession is a helper function for creating a new session. This may look like a lot of code but since we are
// setting up multiple strategies it is a bit longer.
// Usually, you could do:
//
//  session = new(fosite.DefaultSession)
func newSession(user string, req *http.Request) *openid.DefaultSession {
	if user != "" {
		utils.Debug("Creating new session for user: " + user)
	}

	// get hostname from request
	hostname := req.Host

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
		utils.Error("Invalid hostname for OpenID request: " + hostname, nil)
		return nil
	}

	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      hostname,
			Subject:     user,
			ExpiresAt:   time.Now().Add(time.Hour * 6),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
		},
		Headers: &jwt.Headers{
			Extra: map[string]interface{}{
				"kid": "1",
			},
		},
	}
}
