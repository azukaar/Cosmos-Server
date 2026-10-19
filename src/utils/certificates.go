package utils

import (
	"sync"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"time"
	"crypto/rand"
	"crypto"
	"crypto/rsa"
	"crypto/ed25519"
	"crypto/x509/pkix"
	"crypto/ecdsa"
	"os"
	"strings"
	"context"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/registration"
	"github.com/go-acme/lego/v5/challenge/tlsalpn01"
	"github.com/go-acme/lego/v5/providers/dns"
	"github.com/go-acme/lego/v5/challenge/dns01"
)

type CAConfig struct {
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
}

// generateCA creates a new CA certificate if it doesn't exist
func generateCA() (*CAConfig, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	id := GenerateRandomString(5)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cosmos Personal Server CA " + id},
			CommonName:   "Cosmos Root CA " + id,
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(100 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	return &CAConfig{
		Certificate: cert,
		PrivateKey:  privateKey,
	}, nil
}

func loadCA() (*CAConfig, error) {
	config := GetMainConfig()
	CACert := config.HTTPConfig.CACert
	PAPrivateKey := config.HTTPConfig.CAPrivateKey

	if CACert == "" || PAPrivateKey == "" {
		return nil, nil
	}

	certBlock, _ := pem.Decode([]byte(CACert))

	if certBlock == nil {
		return nil, nil
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)

	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode([]byte(PAPrivateKey))

	if keyBlock == nil {
		return nil, nil
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)

	if err != nil {
		return nil, err
	}

	return &CAConfig{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

func saveCA(ca *CAConfig) error {
	baseMainConfig := GetBaseMainConfig()

	baseMainConfig.HTTPConfig.CACert = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Certificate.Raw}))
	baseMainConfig.HTTPConfig.CAPrivateKey = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(ca.PrivateKey)}))

	SetBaseMainConfig(baseMainConfig)
	Log("Saved new CA certificate and private key")

	return nil
}

func GenerateRSAWebCertificates(domains []string) (string, string, error) {
	// Try to load existing CA or generate new one
	ca, err := loadCA()
	if err != nil || ca == nil || ca.Certificate == nil || ca.PrivateKey == nil {
		// If CA doesn't exist, generate and save it
		ca, err = generateCA()
		if err != nil {
			return "", "", err
		}
		if err := saveCA(ca); err != nil {
			return "", "", err
		}
	}

	// Generate web server private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cosmos Personal Server"},
			CommonName:   domains[0], // Use first domain as CN
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().AddDate(0, 0, 364),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:             domains,
		SubjectKeyId:         []byte{1, 2, 3, 4, 6},
		AuthorityKeyId:       ca.Certificate.SubjectKeyId,
	}

	// Create certificate signed by CA
	cert, err := x509.CreateCertificate(rand.Reader, &template, ca.Certificate, &privateKey.PublicKey, ca.PrivateKey)
	if err != nil {
		return "", "", err
	}

	// Return certificate and private key PEM encoded
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return string(certPEM), string(keyPEM), nil
}

func GenerateEd25519Certificates() (string, string) {
	// generate self signed certificate

	// generate private key
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		Fatal("Generating ed25519 Key", err)
	}

	bpriv, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		Fatal("Generating ed25519 private Key", err)
	}

	bpub, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		Fatal("Generating ed25519 public Key", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: bpub})), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bpriv}))
}

type CertUser struct {
	Email        string
	Registration *acme.ExtendedAccount
	key          crypto.Signer
}
func (u *CertUser) GetEmail() string {
	return u.Email
}

func (u CertUser) GetRegistration() *acme.ExtendedAccount {
	return u.Registration
}

func (u *CertUser) GetPrivateKey() crypto.Signer {
	return u.key
}


// process-wide env is the only way to hand credentials to a lego provider, so
// DNS-01 issuances run one at a time with only their own zone's env applied
var dnsChallengeEnvLock sync.Mutex

// desecZone is a minimal deSEC API domain object.
type desecZone struct {
	Name string `json:"name"`
}

// desecRRSet is a minimal deSEC API record-set object.
type desecRRSet struct {
	SubName string   `json:"subname"`
	Type    string   `json:"type"`
	Records []string `json:"records"`
}

// desecAPI performs a deSEC API request with the zone's token. It is kept
// dependency-free (stdlib only) so this patch compiles without go.mod churn.
func desecAPI(ctx context.Context, method, token, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, "https://desec.io/api/v1/"+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("desec API returned %s", resp.Status)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// cleanupStaleChallengeTokens removes leftover _acme-challenge.* TXT records
// from a deSEC zone after a successful order.
//
// Why this exists: lego's deSEC provider does a read-modify-write on the whole
// TXT RRSet. When two runs touch the same _acme-challenge name concurrently
// (possible on stable builds without serialization, or when a failed run left a
// token behind), obsolete tokens pile up. The next pre-check then sees a
// different value than the one just written and reports "did not return the
// expected TXT record" forever, no matter how long the propagation wait.
//
// By the time we get here the order has succeeded and we hold
// dnsChallengeEnvLock, so no other DoLetsEncrypt can be mid-flight. Any
// remaining _acme-challenge TXT in the zone is therefore stale and safe to
// remove; the next renewal recreates what it needs.
func cleanupStaleChallengeTokens(zone *DNSZoneConfig) {
	if zone == nil || !strings.EqualFold(zone.DNSChallengeProvider, "desec") {
		return
	}

	token, ok := zone.DNSChallengeConfig["DESEC_TOKEN"]
	if !ok || strings.TrimSpace(token) == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// find the zone that "owns" the apex of this domain (matches lego's
	// GetResponsible resolution)
	var zones []desecZone
	if err := desecAPI(ctx, http.MethodGet, token,
		"domains/?owns_qname="+url.QueryEscape(strings.TrimSuffix(zone.Zone, ".")), &zones); err != nil {
		Warn("[LETSENCRYPT] could not look up the responsible deSEC zone for " + zone.Zone + ": " + err.Error())
		return
	}
	if len(zones) == 0 {
		return
	}

	var rrsets []desecRRSet
	if err := desecAPI(ctx, http.MethodGet, token,
		"domains/"+zones[0].Name+"/rrsets/", &rrsets); err != nil {
		Warn("[LETSENCRYPT] could not list deSEC records in " + zones[0].Name + ": " + err.Error())
		return
	}

	for _, rr := range rrsets {
		if rr.Type != "TXT" || rr.SubName == "" || rr.SubName == "@" {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(rr.SubName), "_acme-challenge") {
			continue
		}
		if err := desecAPI(ctx, http.MethodDelete, token,
			"domains/"+zones[0].Name+"/rrsets/"+url.PathEscape(rr.SubName)+"/TXT/", nil); err != nil {
			Warn("[LETSENCRYPT] could not remove stale challenge record " + rr.SubName + " in " + zones[0].Name + ": " + err.Error())
			continue
		}
		Log("[LETSENCRYPT] removed stale challenge record " + rr.SubName + " from " + zones[0].Name)
	}
}

// DoLetsEncrypt obtains a certificate for the domains: through the DNS
// provider of the zone when one is given, through HTTP-01/TLS-ALPN-01 otherwise.
func DoLetsEncrypt(domains []string, zone *DNSZoneConfig) (string, string) {
	config := GetMainConfig()
	ctx := context.Background()

	directory := "https://acme-v02.api.letsencrypt.org/directory"
	if os.Getenv("ACME_STAGING") == "true" {
		directory = "https://acme-staging-v02.api.letsencrypt.org/directory"
	}

	// the account of this server, registered once: Let's Encrypt refuses more
	// than a few new accounts per IP, and a fresh one per order used to hit that
	account := LoadACMEAccount(directory)
	keyPEM := ""
	var privateKey *ecdsa.PrivateKey
	var err error
	if account != nil {
		privateKey, err = account.PrivateKey()
	} else {
		privateKey, keyPEM, err = NewACMEAccountKey()
	}
	if err != nil {
		Error("LETSENCRYPT_ECDSA", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	myUser := CertUser{
		Email: config.HTTPConfig.SSLEmail,
		key:   privateKey,
	}
	if account != nil {
		myUser.Registration = &acme.ExtendedAccount{Location: account.URL}
	}

	certConfig := lego.NewConfig(&myUser)
	certConfig.CADirURL = directory

	client, err := lego.NewClient(certConfig)
	if err != nil {
		Error("LETSENCRYPT_NEW", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	if zone != nil {
		dnsChallengeEnvLock.Lock()
		defer dnsChallengeEnvLock.Unlock()

		for key, value := range zone.DNSChallengeConfig {
			key = strings.TrimSpace(key)
			os.Setenv(key, strings.TrimSpace(value))
			defer os.Unsetenv(key)
		}

		provider, err := dns.NewDNSChallengeProviderByName(zone.DNSChallengeProvider)
		if err != nil {
			Error("LETSENCRYPT_DNS", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}

		// use the authoritative nameservers
		resolvers := []string{}

		if zone.DNSChallengeResolvers == "" {
			nameservers, err := net.LookupNS(zone.Zone)
			if err == nil {
				for _, ns := range nameservers {
					resolvers = append(resolvers, ns.Host)
				}
			}

			// append the default resolvers
			resolvers = append(resolvers, "8.8.8.8", "1.1.1.1")
		} else {
			resolvers = strings.Split(zone.DNSChallengeResolvers, ",")
			// trim
			for i, r := range resolvers {
				resolvers[i] = strings.TrimSpace(r)
			}
		}

		propagationWaitSec := zone.DNSChallengePropagationWait
		if propagationWaitSec <= 0 {
			propagationWaitSec = 30
		}
		var PropagationWait = time.Duration(propagationWaitSec) * time.Second

		// lego v5: recursive nameservers moved from a per-challenge option to the shared DNS client
		dns01.SetDefaultClient(dns01.NewClient(&dns01.Options{
			RecursiveNameservers: resolvers,
		}))

		err = client.Challenge.SetDNS01Provider(provider,
			dns01.CondOptions(zone.DisablePropagationChecks,
				dns01.PropagationWait(PropagationWait, true)),
		)
		if err != nil {
			Error("LETSENCRYPT_DNS01", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}
	} else {
		err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", config.HTTPConfig.HTTPPort))
		if err != nil {
			Error("LETSENCRYPT_HTTP01", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}

		err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", config.HTTPConfig.HTTPSPort))
		if err != nil {
			Error("LETSENCRYPT_TLS01", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}
	}

	if account == nil {
		reg, err := client.Registration.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			Error("LETSENCRYPT_REGISTER", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}
		myUser.Registration = reg

		if err := SaveACMEAccount(ACMEAccount{
			Directory: directory,
			URL:       reg.Location,
			Email:     myUser.Email,
			Key:       keyPEM,
		}); err != nil {
			Warn("ACME: the account could not be saved, the next order will register a new one: " + err.Error())
		} else {
			Log("ACME: registered the Let's Encrypt account of this server")
		}
	}

	request := certificate.ObtainRequest{
		Domains: LetsEncryptValidOnly(domains, zone != nil),
		Bundle:  true,
		// lego v5: moved off lego.Config.Certificate, and CN is now opt-in
		KeyType:          certcrypto.RSA2048,
		EnableCommonName: true,
	}
	certificates, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		Error("LETSENCRYPT_OBTAIN", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	// defensive: drop leftovers from failed orders so the next renewal starts
	// from a clean RRSet. Best-effort; errors are logged only and never fail a
	// certificate that was already obtained.
	if zone != nil {
		cleanupStaleChallengeTokens(zone)
	}

	// return cert and key
	return string(certificates.Certificate), string(certificates.PrivateKey)
}

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *acme.ExtendedAccount
	key          crypto.Signer
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *acme.ExtendedAccount {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.Signer {
	return u.key
}