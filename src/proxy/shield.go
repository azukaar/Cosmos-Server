package proxy

import (
	"github.com/azukaar/cosmos-server/src/utils"
	"sync"
	"time"
	"net/http"
	"fmt"
	"net"
	"math"
	"strconv"
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
}

type smartShieldState struct {
	sync.Mutex
	requests []*SmartResponseWriterWrapper
	bans []*userBan
}

type userUsedBudget struct {
	ClientID string
	Time float64
	Requests int
	Bytes int64
	Simultaneous int
}

var shield smartShieldState

func (shield *smartShieldState) GetServerNbReq() int {
	shield.Lock()
	defer shield.Unlock()
	nbRequests := 0

	for i := len(shield.requests) - 1; i >= 0; i-- {
		request := shield.requests[i]
		if(request.IsOld()) {
			return nbRequests
		}
		if(!request.IsOver()) {
			nbRequests++
		}
	}

	return nbRequests
}

func (shield *smartShieldState) GetUserUsedBudgets(ClientID string) userUsedBudget {
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

func (shield *smartShieldState) isAllowedToReqest(policy utils.SmartShieldPolicy, userConsumed userUsedBudget) bool {
	shield.Lock()
	defer shield.Unlock()

	ClientID := userConsumed.ClientID
	
	nbTempBans := 0
	nbStrikes := 0

	// Check for bans
	for i := len(shield.bans) - 1; i >= 0; i-- {
		ban := shield.bans[i]
		if ban.banType == PERM {
			return false
		} else if ban.banType == TEMP {
			if(ban.time.Add(4 * 3600 * time.Second).Before(time.Now())) {
				return false
			} else if (ban.time.Add(72 * 3600 * time.Second).Before(time.Now())) {
				nbTempBans++
			}
		} else if ban.banType == STRIKE {
			return false
			if(ban.time.Add(3600 * time.Second).Before(time.Now())) {
				return false
			} else if (ban.time.Add(24 * 3600 * time.Second).Before(time.Now())) {
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
		return false
	} else if nbStrikes >= 3 {
		// temp ban
		shield.bans = append(shield.bans, &userBan{
			ClientID: ClientID,
			banType: TEMP,
			time: time.Now(),
		})
		return false
	}

	// Check for new strikes
	if (userConsumed.Time > (policy.PerUserTimeBudget * float64(policy.PolicyStrictness))) || 
		 (userConsumed.Requests > (policy.PerUserRequestLimit * policy.PolicyStrictness)) ||
		 (userConsumed.Bytes > (policy.PerUserByteLimit * int64(policy.PolicyStrictness))) ||
		 (userConsumed.Simultaneous > (policy.PerUserSimultaneous * policy.PolicyStrictness)) {
		shield.bans = append(shield.bans, &userBan{
			ClientID: ClientID,
			banType: STRIKE,
			time: time.Now(),
		})
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
	
	if throttle > 0 {
		utils.Debug(fmt.Sprintf("User Time: %f, Requests: %d, Bytes: %d", userConsumed.Time, userConsumed.Requests, userConsumed.Bytes))
		utils.Debug(fmt.Sprintf("Policy Time: %f, Requests: %d, Bytes: %d", policy.PerUserTimeBudget, policy.PerUserRequestLimit, policy.PerUserByteLimit))	
		utils.Debug(fmt.Sprintf("Throttling: %d", throttle))
	}

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
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func isPrivileged(req *http.Request, policy utils.SmartShieldPolicy) bool {
	role, _ := strconv.Atoi(req.Header.Get("x-cosmos-role"))
	return role >= policy.PrivilegedGroups
}

func SmartShieldMiddleware(policy utils.SmartShieldPolicy) func(http.Handler) http.Handler {
	if policy.Enabled == false {
		return func(next http.Handler) http.Handler {
			return next
		}
	} else {
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
			utils.Log("SmartShield: Request received")
			currentGlobalRequests := shield.GetServerNbReq() + 1
			utils.Debug(fmt.Sprintf("SmartShield: Current global requests: %d", currentGlobalRequests))

			if currentGlobalRequests > policy.MaxGlobalSimultaneous && !isPrivileged(r, policy) {
				utils.Log("SmartShield: Too many users on the server")
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			clientID := GetClientID(r)
			userConsumed := shield.GetUserUsedBudgets(clientID)

			if !isPrivileged(r, policy) && !shield.isAllowedToReqest(policy, userConsumed) {
				utils.Log("SmartShield: User is blocked due to abuse")
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
					shield: shield,
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
					shield.Unlock()
				})()

				next.ServeHTTP(wrapper, r)
				close(done)
			}
		})
	}
}