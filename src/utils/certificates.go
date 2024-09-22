package utils

import (
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"time"
	"crypto/rand"
	"crypto"
	"crypto/rsa"
	"crypto/ed25519"
	"crypto/x509/pkix"
	"encoding/asn1"
	"crypto/ecdsa"
	"crypto/elliptic"
	"os"
	"strings"
	"net"
	
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/challenge/dns01"
)

func GenerateRSAWebCertificates(domains []string) (string, string) {
	// generate self signed certificate

	Log("Generating RSA Web Certificates for " + GetMainConfig().HTTPConfig.Hostname)

	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		Fatal("Generating RSA Key", err)
	}

	// generate public key
	publicKey := &privateKey.PublicKey

	// random SerialNumber
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		Fatal("Generating Serial Number", err)
	}

	// generate certificate
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cosmos Personal Server"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,

		DNSNames: domains,

		// IPAddresses: []net.IP{},

		SubjectKeyId: []byte{1, 2, 3, 4, 6},

		AuthorityKeyId: []byte{1, 2, 3, 4, 5},

		PermittedDNSDomainsCritical: false,

		PermittedDNSDomains: domains,

		// PermittedIPRanges: ,

		ExcludedDNSDomains: []string{},

		// ExcludedIPRanges:,

		PermittedURIDomains: []string{},

		ExcludedURIDomains: []string{},

		CRLDistributionPoints: []string{},

		PolicyIdentifiers: []asn1.ObjectIdentifier{},

	}

	// create certificate

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		Fatal("Creating RSA Key", err)
	}

	// return private , and public key

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})), string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}))
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
	Registration *registration.Resource
	key          crypto.PrivateKey
}
func (u *CertUser) GetEmail() string {
	return u.Email
}

func (u CertUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *CertUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}


func DoLetsEncrypt() (string, string) {
	config := GetMainConfig()
	LetsEncryptErrors = []string{}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		Error("LETSENCRYPT_ECDSA", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	myUser := CertUser{
		Email: config.HTTPConfig.SSLEmail,
		key:   privateKey,
	}

	certConfig := lego.NewConfig(&myUser)
	
	domains := GetAllHostnames(true, true)

	if os.Getenv("ACME_STAGING") == "true" {
		certConfig.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	} else {
		certConfig.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	}

	certConfig.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(certConfig)
	if err != nil {
		Error("LETSENCRYPT_NEW", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	if config.HTTPConfig.DNSChallengeProvider != "" {
		provider, err := dns.NewDNSChallengeProviderByName(config.HTTPConfig.DNSChallengeProvider)
		if err != nil {
			Error("LETSENCRYPT_DNS", err)
			LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
			return "", ""
		}

		// use the authoritative nameservers
		resolvers := []string{}

		if config.HTTPConfig.DNSChallengeResolvers == "" {
			processedDomains := map[string]bool{}

			for _, domain := range domains {
				levels := strings.Split(domain, ".")
				if len(levels) >= 2 {
					tld := strings.Join(levels[len(levels)-2:], ".")
					if processedDomains[tld] {
						continue
					}

					nameservers, err := net.LookupNS(tld)
					
					if err != nil {
						continue
					}
					
					for _, ns := range nameservers {
						resolvers = append(resolvers, ns.Host)
					}

					processedDomains[tld] = true
				}
			}

			// append the default resolvers
			resolvers = append(resolvers, "8.8.8.8", "1.1.1.1")
		} else {
			resolvers = strings.Split(config.HTTPConfig.DNSChallengeResolvers, ",")
			// trim
			for i, r := range resolvers {
				resolvers[i] = strings.TrimSpace(r)
			}
		}

		err = client.Challenge.SetDNS01Provider(provider,
			dns01.AddRecursiveNameservers(resolvers))
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

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		Error("LETSENCRYPT_REGISTER", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: LetsEncryptValidOnly(domains, config.HTTPConfig.DNSChallengeProvider != ""),
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		Error("LETSENCRYPT_OBTAIN", err)
		LetsEncryptErrors = append(LetsEncryptErrors, err.Error())
		return "", ""
	}

	// return cert and key
	return string(certificates.Certificate), string(certificates.PrivateKey)
}

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}