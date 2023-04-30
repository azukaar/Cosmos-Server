package utils

import (
	"context"
	"net/http"
	"time"
	"net"

	"github.com/mxk/go-flowrate/flowrate"
	"github.com/oschwald/geoip2-golang"
)

// https://github.com/go-chi/chi/blob/master/middleware/timeout.go

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
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Served-By-Cosmos", "1")

		next.ServeHTTP(w, r)
	})
}

func CORSHeader(origin string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

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
func BlockByCountryMiddleware(blockedCountries []string) func(http.Handler) http.Handler {
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

				for _, blockedCountry := range blockedCountries {
					if config.ServerCountry != countryCode && countryCode == blockedCountry {
						http.Error(w, "Access denied", http.StatusForbidden)
						return
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
				http.Error(w, "Bad Request: Invalid request.", http.StatusBadRequest)
				return
			}
		}

		// If it's not a POST request or the POST request has a Referer header, pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}