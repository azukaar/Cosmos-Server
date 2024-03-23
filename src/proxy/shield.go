package proxy

import (
	"sync"
	"time"
	"net/http"
	"fmt"
	"net"
	"math"
	"strconv"

	"github.com/azukaar/cosmos-server/src/utils"
	"github.com/azukaar/cosmos-server/src/metrics"
)

/*
	TODO :
	 - Recalculate throttle every gb for writer wrapper?
*/

const (
	STRIKE = 0
	TEMP = 1
	PERM = 2
)
type userBan struct {
	ClientID string
	banType int
	time time.Time
	reason string
	shieldID string
}

type smartShieldState struct {
	sync.Mutex
	requests []*SmartResponseWriterWrapper
	bans []*userBan
}

type userUsedBudget struct {
	ClientID string `json:"clientID"`
	Time float64 `json:"time"`
	Requests int `json:"requests"`
	Bytes int64 `json:"bytes"`
	Simultaneous int `json:"simultaneous"`
}

var shield smartShieldState

func GetShield() int {
	return len(shield.requests) + len(shield.bans)
}

func CleanUp() {
	shield.Lock()
	defer shield.Unlock()

	shieldSize := len(shield.requests) + len(shield.bans)

	for i := len(shield.requests) - 1; i >= 0; i-- {
		request := shield.requests[i]
		if(request.IsOld()) {
			shield.requests = append(shield.requests[:i], shield.requests[i+1:]...)
		}
	}

	for i := len(shield.bans) - 1; i >= 0; i-- {
		ban := shield.bans[i]
		if(ban.banType == TEMP && ban.time.Add(72 * 3600 * time.Second).Before(time.Now())) {
			shield.bans = append(shield.bans[:i], shield.bans[i+1:]...)
		}
		if(ban.banType == STRIKE && ban.time.Add(72 * 3600 * time.Second).Before(time.Now())) {
			shield.bans = append(shield.bans[:i], shield.bans[i+1:]...)
		}
	}

	utils.Log("SmartShield: Cleaned up " + fmt.Sprintf("%d", shieldSize - (len(shield.requests) + len(shield.bans))) + " items")
}

func (shield *smartShieldState) GetServerNbReq(shieldID string) int {
	shield.Lock()
	defer shield.Unlock()
	nbRequests := 0

	for i := len(shield.requests) - 1; i >= 0; i-- {
		request := shield.requests[i]
		if(request.IsOld()) {
			return nbRequests
		}
		if(!request.IsOver() && request.shieldID == shieldID) {
			nbRequests++
		}
	}

	return nbRequests
}

func (shield *smartShieldState) GetUserUsedBudgets(shieldID string, ClientID string) userUsedBudget {
	shield.Lock()
	defer shield.Unlock()

	userConsumed	:= userUsedBudget{
		ClientID: ClientID,
		Time: 0,
		Requests: 0,
		Bytes: 0,
		Simultaneous: 0,
	}

	// Check for recent requests
	for i := len(shield.requests) - 1; i >= 0; i-- {
		request := shield.requests[i]
		
		if(request.shieldID != shieldID) {
			continue
		}

		if(request.IsOld()) {
			return userConsumed
		}
		if request.ClientID == ClientID && !request.IsOld() {
			if(request.IsOver()) {
				userConsumed.Time += request.TimeEnded.Sub(request.TimeStarted).Seconds()
			} else {
				userConsumed.Time += time.Now().Sub(request.TimeStarted).Seconds()
				userConsumed.Simultaneous++
			}
			userConsumed.Requests += request.RequestCost
			userConsumed.Bytes += request.Bytes
		}
	}

	return userConsumed
}

func (shield *smartShieldState) GetLastBan(policy utils.SmartShieldPolicy, userConsumed userUsedBudget) *userBan {
	shield.Lock()
	defer shield.Unlock()

	ClientID := userConsumed.ClientID

	// Check for bans
	for i := len(shield.bans) - 1; i >= 0; i-- {
		ban := shield.bans[i]
		if ban.banType == STRIKE && ban.ClientID == ClientID {
			return ban
		}
	}

	return nil
}

func (shield *smartShieldState) isAllowedToReqest(shieldID string, policy utils.SmartShieldPolicy, userConsumed userUsedBudget) bool {
	shield.Lock()
	defer shield.Unlock()

	ClientID := userConsumed.ClientID

	if ClientID == "192.168.1.1" || 
		 ClientID == "192.168.0.1" ||
		 ClientID == "192.168.0.254" ||
		 ClientID == "172.17.0.1" {
		return true
	}
	
	nbTempBans := 0
	nbStrikes := 0

	// Check for bans
	for i := len(shield.bans) - 1; i >= 0; i-- {
		ban := shield.bans[i]

		if ban.banType == PERM && ban.ClientID == ClientID {
			return false
		} else if ban.banType == TEMP && ban.ClientID == ClientID {
			if(ban.time.Add(4 * 3600 * time.Second).After(time.Now())) {
				return false
			} else if (ban.time.Add(72 * 3600 * time.Second).After(time.Now())) {
				nbTempBans++
			}
		} else if ban.banType == STRIKE && ban.ClientID == ClientID {
			if(ban.time.Add(3600 * time.Second).After(time.Now())) {
				return false
			} else if (ban.time.Add(24 * 3600 * time.Second).After(time.Now())) {
				nbStrikes++
			}
		}
	}

	// Check for new bans
	if nbTempBans >= 3 {
		// perm ban
		shield.bans = append(shield.bans, &userBan{
			ClientID: ClientID,
			banType: PERM,
			time: time.Now(),
		})

		utils.Warn("User " + ClientID + " has been banned permanently: "+ fmt.Sprintf("%+v", userConsumed))
		return false
	} else if nbStrikes >= 3 {
		// temp ban
		shield.bans = append(shield.bans, &userBan{
			ClientID: ClientID,
			banType: TEMP,
			time: time.Now(),
		})
		utils.Warn("User " + ClientID + " has been banned temporarily: "+ fmt.Sprintf("%+v", userConsumed))
		return false
	}

	// Check for new strikes
	if (userConsumed.Time > (policy.PerUserTimeBudget * float64(policy.PolicyStrictness))) || 
		 (userConsumed.Requests > (policy.PerUserRequestLimit * policy.PolicyStrictness)) ||
		 (userConsumed.Bytes > (policy.PerUserByteLimit * int64(policy.PolicyStrictness))) ||
		 (userConsumed.Simultaneous > (policy.PerUserSimultaneous * policy.PolicyStrictness * 15)) {
		shield.bans = append(shield.bans, &userBan{
			ClientID: ClientID,
			banType: STRIKE,
			time: time.Now(),
			reason: fmt.Sprintf("%+v out of %+v", userConsumed, policy),
			shieldID: shieldID,
		})
		utils.Warn("User " + ClientID + " has received a strike: "+ fmt.Sprintf("%+v", userConsumed))
		return false
	}

	return true
}

func (shield *smartShieldState) computeThrottle(policy utils.SmartShieldPolicy, userConsumed userUsedBudget) int {	
	shield.Lock()
	defer shield.Unlock()

	throttle := 0

	overReq := policy.PerUserRequestLimit - userConsumed.Requests
	overReqRatio := float64(overReq) / float64(policy.PerUserRequestLimit)
	if overReq < 0 {
		newThrottle := int(float64(2500) * -overReqRatio)
		if newThrottle > throttle {
			throttle = newThrottle
		}
	}
	
	overByte := policy.PerUserByteLimit - userConsumed.Bytes
	overByteRatio := float64(overByte) / float64(policy.PerUserByteLimit)
	if overByte < 0 {
		newThrottle := int(float64(40) * -overByteRatio)
		if newThrottle > throttle {
			throttle = newThrottle
		}
	}

	overSim := policy.PerUserSimultaneous - userConsumed.Simultaneous
	overSimRatio := float64(overSim) / float64(policy.PerUserSimultaneous)
	if overSim < 0 {
		newThrottle := int(float64(50) * -overSimRatio)
		if newThrottle > throttle {
			throttle = newThrottle
		}
	}
	
	// if throttle > 0 {
		utils.Debug(fmt.Sprintf("User Time: %f, Requests: %d, Bytes: %d", userConsumed.Time, userConsumed.Requests, userConsumed.Bytes))
		utils.Debug(fmt.Sprintf("Policy Time: %f, Requests: %d, Bytes: %d", policy.PerUserTimeBudget, policy.PerUserRequestLimit, policy.PerUserByteLimit))	
		utils.Debug(fmt.Sprintf("Throttling: %d", throttle))
	// }

	return throttle
}

func calculateLowestExhaustedPercentage(policy utils.SmartShieldPolicy, userConsumed userUsedBudget) int64 {
	timeExhaustedPercentage := 100 - (userConsumed.Time / policy.PerUserTimeBudget) * 100
	requestsExhaustedPercentage := 100 - (float64(userConsumed.Requests) / float64(policy.PerUserRequestLimit)) * 100
	bytesExhaustedPercentage := 100 - (float64(userConsumed.Bytes) / float64(policy.PerUserByteLimit)) * 100

	// utils.Debug(fmt.Sprintf("Time: %f, Requests: %d, Bytes: %d", timeExhaustedPercentage, requestsExhaustedPercentage, bytesExhaustedPercentage))
	
	return int64(math.Max(0, math.Min(math.Min(timeExhaustedPercentage, requestsExhaustedPercentage), bytesExhaustedPercentage)))
}

func GetClientID(r *http.Request) string {
	// when using Docker we need to get the real IP
	UseForwardedFor := utils.GetMainConfig().HTTPConfig.UseForwardedFor

	if UseForwardedFor && r.Header.Get("x-forwarded-for") != "" {
		ip, _, _ := net.SplitHostPort(r.Header.Get("x-forwarded-for"))
		utils.Debug("SmartShield: Getting client ID " + ip)
		return ip
	} else {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		utils.Debug("SmartShield: Getting client ID " + ip)
		return ip
	}
}

func isPrivileged(req *http.Request, policy utils.SmartShieldPolicy) bool {
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	return role >= policy.PrivilegedGroups
}

func SmartShieldMiddleware(shieldID string, route utils.ProxyRouteConfig) func(http.Handler) http.Handler {
	policy := route.SmartShield

	if policy.Enabled {
		if(policy.PerUserTimeBudget == 0) {
			policy.PerUserTimeBudget = 2 * 60 * 60 * 1000 // 2 hours
		}
		if(policy.PerUserRequestLimit == 0) {
			policy.PerUserRequestLimit = 6000 // 100 requests per minute
		}
		if(policy.PerUserByteLimit == 0) {
			policy.PerUserByteLimit = 150 * 1024 * 1024 * 1024 // 150GB
		}
		if(policy.PolicyStrictness == 0) {
			policy.PolicyStrictness = 2 // NORMAL
		}
		if(policy.PerUserSimultaneous == 0) {
			policy.PerUserSimultaneous = 24
		}
		if(policy.MaxGlobalSimultaneous == 0) {
			policy.MaxGlobalSimultaneous = 250
		}
		if(policy.PrivilegedGroups == 0) {
			policy.PrivilegedGroups = utils.ADMIN
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientID := GetClientID(r)

			wrapper := &SmartResponseWriterWrapper {
				ResponseWriter: w,
				ThrottleNext:   0,
				TimeStarted:    time.Now(),
				ClientID:       clientID,
				RequestCost:    1,
				Method: 				r.Method,
				shield: &shield,
				shieldID: shieldID,
				policy: policy,
				isPrivileged: isPrivileged(r, policy),
			}

			if !policy.Enabled {
				next.ServeHTTP(wrapper, r)
				wrapper.TimeEnded = time.Now()
				wrapper.isOver = true

				statusText := "success"
				level := "info"
				if wrapper.Status >= 400 {
					statusText = "error"
					level = "warning"
				}

				utils.TriggerEvent(
					"cosmos.proxy.response." + route.Name + "." + statusText,
					"Proxy Response " + route.Name + " " + statusText,
					level,
					"route@" + route.Name,
					map[string]interface{}{
					"route": route.Name,
					"status": wrapper.Status,
					"method": wrapper.Method,
					"clientID": wrapper.ClientID,
					"time": wrapper.TimeEnded.Sub(wrapper.TimeStarted).Seconds(),
					"bytes": wrapper.Bytes,
					"url": r.URL,
				})

				go metrics.PushRequestMetrics(route, wrapper.Status, wrapper.TimeStarted, wrapper.Bytes)

				return
			}

			currentGlobalRequests := shield.GetServerNbReq(shieldID) + 1
			utils.Debug(fmt.Sprintf("SmartShield: Current global requests: %d", currentGlobalRequests))

			if !isPrivileged(r, policy) {
				tooManyReq := currentGlobalRequests > policy.MaxGlobalSimultaneous
				wayTooManyReq := currentGlobalRequests > policy.MaxGlobalSimultaneous * 10
				retries := 50
				if wayTooManyReq {
					go metrics.PushShieldMetrics("smart-shield")
					utils.Log("SmartShield: WAYYYY Too many users on the server. Aborting right away.")
					http.Error(w, "Too many requests", http.StatusTooManyRequests)
					return
				}
				for tooManyReq {
					time.Sleep(5000 * time.Millisecond)
					currentGlobalRequests := shield.GetServerNbReq(shieldID) + 1
					tooManyReq = currentGlobalRequests > policy.MaxGlobalSimultaneous
					retries--
					if retries <= 0 {
						go metrics.PushShieldMetrics("smart-shield")
						utils.Log("SmartShield: Too many users on the server")
						http.Error(w, "Too many requests", http.StatusTooManyRequests)
						return
					}
				}
			}

			userConsumed := shield.GetUserUsedBudgets(shieldID, clientID)

			if !isPrivileged(r, policy) && !shield.isAllowedToReqest(shieldID, policy, userConsumed) {
				lastBan := shield.GetLastBan(policy, userConsumed)
				go metrics.PushShieldMetrics("smart-shield")
				utils.IncrementIPAbuseCounter(clientID)

				utils.TriggerEvent(
					"cosmos.proxy.shield.abuse." + route.Name,
					"Proxy Shield " + route.Name + " Abuse by " + clientID,
					"warning",
					"route@" + route.Name,
					map[string]interface{}{
					"route": route.Name,
					"consumed": userConsumed,
					"lastBan": lastBan,
					"clientID": clientID,
					"url": r.URL,
				})

				utils.Log("SmartShield: User is blocked due to abuse: " + fmt.Sprintf("%+v", lastBan))
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			} else {
				throttle := 0
				if(!isPrivileged(r, policy)) {
					throttle = shield.computeThrottle(policy, userConsumed)
				}

				wrapper := &SmartResponseWriterWrapper {
					ResponseWriter: w,
					ThrottleNext:   throttle,
					TimeStarted:    time.Now(),
					ClientID:       clientID,
					RequestCost:    1,
					Method: 				r.Method,
					shield: &shield,
					shieldID: shieldID,
					policy: policy,
					isPrivileged: isPrivileged(r, policy),
				}

				// add rate limite headers
				In20Minutes := strconv.FormatInt(time.Now().Add(20 * time.Minute).Unix(), 10)
				w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(calculateLowestExhaustedPercentage(policy, userConsumed), 10))
				w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(int64(policy.PerUserRequestLimit), 10))
				w.Header().Set("X-RateLimit-Reset", In20Minutes)

				shield.Lock()
				shield.requests = append(shield.requests, wrapper)
				shield.Unlock()
				
				ctx := r.Context()
				done := make(chan struct{})

				go (func() {
					select {
					case <-ctx.Done():
					case <-done:
					}	
					shield.Lock()
					wrapper.TimeEnded = time.Now()
					wrapper.isOver = true

					statusText := "success"
					level := "info"
					if wrapper.Status >= 400 {
						statusText = "error"
						level = "warning"
					}

					utils.TriggerEvent(
						"cosmos.proxy.response." + route.Name + "." + statusText,
						"Proxy Response " + route.Name + " " + statusText,
						level,
						"route@" + route.Name,
						map[string]interface{}{
						"route": route.Name,
						"status": wrapper.Status,
						"method": wrapper.Method,
						"clientID": wrapper.ClientID,
						"time": wrapper.TimeEnded.Sub(wrapper.TimeStarted).Seconds(),
						"bytes": wrapper.Bytes,
						"url": r.URL,
					})

					go metrics.PushRequestMetrics(route, wrapper.Status, wrapper.TimeStarted, wrapper.Bytes)
					shield.Unlock()
				})()

				next.ServeHTTP(wrapper, r)
				close(done)
			}
		})
	}
}