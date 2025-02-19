package authorizationserver

import (
	"math/rand"
	"crypto/rsa"
	"github.com/ory/fosite"
	"time"
	"net/http"
	"os"
	"strings"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"
	"encoding/json"
	"net/url"
	"net/http/httptest"
	"go.mongodb.org/mongo-driver/mongo"
	
	// "fmt"

	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"

	"github.com/azukaar/cosmos-server/src/utils"
	userMod "github.com/azukaar/cosmos-server/src/user"

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
		SendDebugMessagesToClients: config.LoggingLevel == "DEBUG",
	}

	store := storage.NewMemoryStore()
	
	// loop config.OpenIDClients
	for _, client := range config.OpenIDClients {
		utils.Log("Registering OpenID client: " + client.ID)

		// register client
		store.Clients[client.ID] = &fosite.DefaultClient{
			ID:             client.ID,
			Secret:         []byte(client.Secret),
			RedirectURIs:   strings.Split(client.Redirect, ","),
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

	// Add proxy route clients
	for _, route := range config.HTTPConfig.ProxyConfig.Routes {
		if route.AuthEnabled && route.UseHost && !route.Disabled {
			utils.Log("Registering OpenID client for route: " + route.Host)
			client := utils.GetProxyOIDCredentials(route, true)
			store.Clients[client.ID] = client
		}
	}

	utils.Log("OpenID server initialized")
}

func RegisterHandlersDetect(wellKnown *mux.Router, srapi *mux.Router) {
	wellKnown.HandleFunc("/.well-known/openid-detect", discoverEndpoint)
	srapi.HandleFunc("/oauth2/detect-callback", detectCallbackEndpoint)
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
	// set well-known endpoints to be json encoded
	wellKnown.Use(utils.AcceptHeader("application/json"))
	
	wellKnown.HandleFunc("/.well-known/openid-configuration", discoverEndpoint)
	wellKnown.HandleFunc("/.well-known/jwks.json", jwksEndpoint)
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

func detectCallbackEndpoint(w http.ResponseWriter, req *http.Request) {
	config := utils.GetMainConfig()

	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")
	parts := strings.Split(state, ",,")
	if len(parts) != 2 {
		utils.Error("Invalid state format", nil)
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	hashStr, encodedPath := parts[0], parts[1]
	path, err := url.QueryUnescape(encodedPath)
	if err != nil {
		utils.Error("Failed to decode path", err)
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Verify hash matches path
	expectedHash := sha256.Sum256([]byte(path + config.HTTPConfig.AuthPrivateKey[32:64]))
	expectedHashStr := base64.RawURLEncoding.EncodeToString(expectedHash[:16])
	if hashStr != expectedHashStr {
		utils.Error("Invalid state hash", nil)
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return		
	}

	
	if code == "" {
			error := req.URL.Query().Get("error")
			errorDesc := req.URL.Query().Get("error_description")
			utils.Error("OAuth callback error: "+error+" - "+errorDesc, nil)
			http.Error(w, "OAuth callback error: "+error+" - "+errorDesc, http.StatusInternalServerError)
			return
	}

	_, route := utils.FindRouteByReqHost(req.Host)
	client := utils.GetProxyOIDCredentials(*route, false)
	strcli := (string)(client.Secret)
	
	utils.Log("OpenID Direct: Exchanging code for token for client: " + client.ID)

	// Create form data - remove client_id and client_secret from body
	formData := url.Values{
			"grant_type":   {"authorization_code"},
			"code":         {code},
			"redirect_uri": {client.RedirectURIs[0]},
	}

	tokenReq, err := http.NewRequestWithContext(
			req.Context(),
			"POST", 
			"https://" + config.HTTPConfig.Hostname + "/oauth2/token",
			strings.NewReader(formData.Encode()),
	)

	if err != nil {
			utils.Error("Failed to create token request", err)
			http.Error(w, "Failed to create token request", http.StatusInternalServerError)
			return
	}

	// Set Basic auth header for client authentication
	tokenReq.SetBasicAuth(client.ID, strcli)
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set additional required attributes
	tokenReq.RemoteAddr = config.HTTPConfig.Hostname
	tokenReq.Host = config.HTTPConfig.Hostname
	tokenReq.TLS = req.TLS
	
	tokenRW := httptest.NewRecorder()
	tokenEndpoint(tokenRW, tokenReq)
	
	resp := tokenRW.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			utils.Error("Token endpoint returned non-200 status: "+string(body), nil)
			http.Error(w, "Token endpoint returned non-200 status: "+string(body), http.StatusInternalServerError)
			return
	}

	var tokenResp struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			IDToken     string `json:"id_token"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			utils.Error("Failed to decode token response", err)
			http.Error(w, "Failed to decode token response", http.StatusInternalServerError)
			return
	}

	// encode json response token
	idToken := tokenResp.IDToken

	// decode id token
	claims := jwt.MapClaims{}

	_, err = jwt.ParseWithClaims(idToken, claims, func(token *jwt.Token) (interface{}, error) {
			return AuthPrivateKey.Public(), nil
	})

	if err != nil {
			utils.Error("Failed to decode id token", err)
			http.Error(w, "Failed to decode id token", http.StatusInternalServerError)
			return
	}

	// get user from claims
	user := claims["sub"].(string)

	c, closeDb, errCo := utils.GetEmbeddedCollection(utils.GetRootAppId(), "users")
	defer closeDb()
	if errCo != nil {
		utils.Error("Database Connect", errCo)
		utils.HTTPError(w, "Database Error", http.StatusInternalServerError, "DB001")
		return
	}

	nickname := utils.Sanitize(user)

	userInBase := utils.User{}

	utils.Debug("UserLogin: Logging user " + nickname)

	err3 := c.FindOne(nil, map[string]interface{}{
		"Nickname": nickname,
	}).Decode(&userInBase)

	if err3 == mongo.ErrNoDocuments {
		utils.Error("UserLogin: User not found", err3)
		utils.HTTPError(w, "User Logging Error", http.StatusInternalServerError, "UL001")
		return
	} else if userInBase.Password == "" {
		utils.Error("UserLogin: User not registered", nil)
		utils.HTTPError(w, "User not registered", http.StatusInternalServerError, "UL002")
		return
	}
	
	userMod.SendUserToken(w, req, userInBase, true, utils.GUEST)

	// redirect to user page 
	// TODO: redirect to RIGHT PAGE
	http.Redirect(w, req, path, http.StatusFound)
}