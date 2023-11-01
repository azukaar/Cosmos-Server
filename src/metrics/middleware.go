package metrics

import (
	"net/http"
	"fmt"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

// responseWriter wraps the original http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func MetricsMiddleware(route utils.ProxyRouteConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Call the next handler (which can be another middleware or the final handler).
		wrappedWriter := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(wrappedWriter, r)

		// Calculate and log the response time.
		responseTime := time.Since(startTime)

		utils.Debug(fmt.Sprintf("[%s] %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, responseTime))

		if !utils.GetMainConfig().MonitoringDisabled {
			go func() {
				if wrappedWriter.status >= 400 {
					PushSetMetric("proxy.all.error", 1, DataDef{
						Max: 0,
						Period: time.Second * 30,
						Label: "Global Request Errors",
						AggloType: "sum",
						SetOperation: "sum",
					})
					PushSetMetric("proxy.route.error."+route.Name, 1, DataDef{
						Max: 0,
						Period: time.Second * 30,
						Label: "Request Errors " + route.Name,
						AggloType: "sum",
						SetOperation: "sum",
					})
				} else {
					PushSetMetric("proxy.all.success", 1, DataDef{
						Max: 0,
						Period: time.Second * 30,
						Label: "Global Request Success",
						AggloType: "sum",
						SetOperation: "sum",
					})
					PushSetMetric("proxy.route.success."+route.Name, 1, DataDef{
						Max: 0,
						Period: time.Second * 30,
						Label: "Request Success " + route.Name,
						AggloType: "sum",
						SetOperation: "sum",
					})
				}

				PushSetMetric("proxy.all.time", int(responseTime.Milliseconds()), DataDef{
					Max: 0,
					Period: time.Second * 30,
					Label: "Global Response Time",
					AggloType: "avg",
					SetOperation: "max",
					Unit: "ms",
				})

				PushSetMetric("proxy.route.time."+route.Name, int(responseTime.Milliseconds()), DataDef{
					Max: 0,
					Period: time.Second * 30,
					Label: "Response Time " + route.Name,
					AggloType: "avg",
					SetOperation: "max",
					Unit: "ms",
				})
			}()
		}

	})
}
}