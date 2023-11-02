package utils

import (
	"context"
	"net/http"
	"time"
	"net"
	"strings"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/mxk/go-flowrate/flowrate"
	"github.com/oschwald/geoip2-golang"
)

// https://github.com/go-chi/chi/blob/master/middleware/timeout.go

var PushShieldMetrics func(string)

type safeInt struct {
	val int64
}

var BannedIPs = sync.Map{}

// Close connection right away if banned (save resources)

func IncrementIPAbuseCounter(ip string) {
	// Load or store a new *safeInt
	actual, _ := BannedIPs.LoadOrStore(ip, &safeInt{})
	counter := actual.(*safeInt)

	// Increment the counter using atomic for concurrent access
	atomic.AddInt64(&counter.val, 1)
}

func getIPAbuseCounter(ip string) int64 {
	// Load the *safeInt
	actual, ok := BannedIPs.Load(ip)
	if !ok {
			return 0
	}
	counter := actual.(*safeInt)

	// Load the value using atomic for concurrent access
	return atomic.LoadInt64(&counter.val)
}

func BlockBannedIPs(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil {
					if hj, ok := w.(http.Hijacker); ok {
							conn, _, err := hj.Hijack()
							if err == nil {
									conn.Close()
							}
					}
					return
        }

				nbAbuse := getIPAbuseCounter(ip)

				// Debug("IP " + ip + " has " + fmt.Sprintf("%d", nbAbuse) + " abuse(s)")

        if nbAbuse > 1000 {
					if hj, ok := w.(http.Hijacker); ok {
							conn, _, err := hj.Hijack()
							if err == nil {
									conn.Close()
							}
					}
					return
				}

        next.ServeHTTP(w, r)
    })
}

func MiddlewareTimeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					Error("Request Timeout. Cancelling.", ctx.Err())
					HTTPError(w, "Gateway Timeout", 
						http.StatusGatewayTimeout, "HTTP002")
					return 
				}
			}()

			w.Header().Set("X-Timeout-Duration", timeout.String())

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

type responseWriter struct {
	http.ResponseWriter
	*flowrate.Writer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func BandwithLimiterMiddleware(max int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if(max > 0) {
				fw := flowrate.NewWriter(w, max)
				w = &responseWriter{w, fw}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func SetSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if(IsHTTPS) {
			// TODO: Add preload if we have a valid certificate
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
		
		w.Header().Set("X-Served-By-Cosmos", "1")
		
		next.ServeHTTP(w, r)
	})
}

func CORSHeader(origin string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AcceptHeader(accept string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", accept)

			next.ServeHTTP(w, r)
		})
	}
}

// GetIPLocation returns the ISO country code for a given IP address.
func GetIPLocation(ip string) (string, error) {
	geoDB, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		return "", err
	}
	defer geoDB.Close()

	parsedIP := net.ParseIP(ip)
	record, err := geoDB.Country(parsedIP)
	if err != nil {
		return "", err
	}

	return record.Country.IsoCode, nil
}

// BlockByCountryMiddleware returns a middleware function that blocks requests from specified countries.
func BlockByCountryMiddleware(blockedCountries []string, CountryBlacklistIsWhitelist bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			countryCode, err := GetIPLocation(ip)

			if err == nil {
				if countryCode == "" {
					Debug("Country code is empty")
				} else {
					Debug("Country code: " + countryCode)
				}

				config := GetMainConfig()

				if CountryBlacklistIsWhitelist {
					if countryCode != "" {
						blocked := true
						for _, blockedCountry := range blockedCountries {
							if config.ServerCountry != countryCode && countryCode == blockedCountry {
								blocked = false
							}
						}

						if blocked {
							PushShieldMetrics("geo")
							IncrementIPAbuseCounter(ip)
							http.Error(w, "Access denied", http.StatusForbidden)
							return
						}
					} else {
						Warn("Missing geolocation information to block IPs")
					}
				} else {
					for _, blockedCountry := range blockedCountries {
						if config.ServerCountry != countryCode && countryCode == blockedCountry {
							http.Error(w, "Access denied", http.StatusForbidden)
							return
						}
					}
				}
			} else {
				Warn("Missing geolocation information to block IPs")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// blockPostWithoutReferer blocks POST requests without a Referer header
func BlockPostWithoutReferer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" || r.Method == "DELETE" {
			referer := r.Header.Get("Referer")
			if referer == "" {
				PushShieldMetrics("referer")
				Error("Blocked POST request without Referer header", nil)
				http.Error(w, "Bad Request: Invalid request.", http.StatusBadRequest)

				ip, _, _ := net.SplitHostPort(r.RemoteAddr)
				if ip != "" {
					IncrementIPAbuseCounter(ip)
				}

				return
			}
		}

		// If it's not a POST request or the POST request has a Referer header, pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

func EnsureHostname(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Debug("Ensuring origin for requested resource from : " + r.Host)

		og := GetMainConfig().HTTPConfig.Hostname
		ni := GetMainConfig().NewInstall

		if ni || og == "0.0.0.0" {
			next.ServeHTTP(w, r)
			return
		}

		hostnames := GetAllHostnames(false, false)

		reqHostNoPort := strings.Split(r.Host, ":")[0]

		isOk := false
		for _, hostname := range hostnames {
			hostnameNoPort := strings.Split(hostname, ":")[0]
			if reqHostNoPort == hostnameNoPort {
				isOk = true
			}
		}

		if !isOk {
			PushShieldMetrics("hostname")
			Error("Invalid Hostname " + r.Host + " for request. Expecting one of " + fmt.Sprintf("%v", hostnames), nil)
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "Bad Request: Invalid hostname. Use your domain instead of your IP to access your server. Check logs if more details are needed.", http.StatusBadRequest)
			
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip != "" {
				IncrementIPAbuseCounter(ip)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}

func IsValidHostname(hostname string) bool {
	og := GetMainConfig().HTTPConfig.Hostname
	ni := GetMainConfig().NewInstall

	if ni || og == "0.0.0.0" {
		return true
	}

	hostnames := GetAllHostnames(false, false)

	reqHostNoPort := strings.Split(hostname, ":")[0]

	reqHostNoPortNoSubdomain := ""

	if parts := strings.Split(reqHostNoPort, "."); len(parts) < 2 {
		reqHostNoPortNoSubdomain = reqHostNoPort
	} else {
		reqHostNoPortNoSubdomain = parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	for _, hostname := range hostnames {
		hostnameNoPort := strings.Split(hostname, ":")[0]
		hostnameNoPortNoSubdomain := ""

		if parts := strings.Split(hostnameNoPort, "."); len(parts) < 2 {
			hostnameNoPortNoSubdomain = hostnameNoPort
		} else {
			hostnameNoPortNoSubdomain = parts[len(parts)-2] + "." + parts[len(parts)-1]
		}

		if reqHostNoPortNoSubdomain == hostnameNoPortNoSubdomain {
			return true
		}
	}

	return false
}

func IPInRange(ipStr, cidrStr string) (bool, error) {
	_, cidrNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false, fmt.Errorf("parse CIDR range: %w", err)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("parse IP: invalid IP address")
	}

	return cidrNet.Contains(ip), nil
}

func Restrictions(RestrictToConstellation bool, WhitelistInboundIPs []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		isUsingWhiteList := len(WhitelistInboundIPs) > 0

		isInWhitelist := false
		isInConstellation := strings.HasPrefix(ip, "192.168.201.") || strings.HasPrefix(ip, "192.168.202.")

		for _, ipRange := range WhitelistInboundIPs {
			Debug("Checking if " + ip + " is in " + ipRange)
			if strings.Contains(ipRange, "/") {
				if ok, _ := IPInRange(ip, ipRange); ok {
					isInWhitelist = true
				}
			} else {
				if ip == ipRange {
					isInWhitelist = true
				}
			}
		}

		if(RestrictToConstellation) {
			if(!isInConstellation) {
				if(!isUsingWhiteList) {
					PushShieldMetrics("ip-whitelists")
					IncrementIPAbuseCounter(ip)
					Error("Request from " + ip + " is blocked because of restrictions", nil)
					Debug("Blocked by RestrictToConstellation isInConstellation isUsingWhiteList")
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				} else if (!isInWhitelist) {
					PushShieldMetrics("ip-whitelists")
					IncrementIPAbuseCounter(ip)
					Error("Request from " + ip + " is blocked because of restrictions", nil)
					Debug("Blocked by RestrictToConstellation isInConstellation isInWhitelist")
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
			}
		} else if(isUsingWhiteList && !isInWhitelist) {
			PushShieldMetrics("ip-whitelists")
			IncrementIPAbuseCounter(ip)
			Error("Request from " + ip + " is blocked because of restrictions", nil)
			Debug("Blocked by RestrictToConstellation isInConstellation isUsingWhiteList isInWhitelist")
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
		})
	}
}
