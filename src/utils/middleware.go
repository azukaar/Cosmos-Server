package utils

import (
	"context"
	"net/http"
	"time"
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
		w.Header().Set("Referrer-Policy", "no-referrer")

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
