package limiter

import (
	"context"
	"golang.org/x/time/rate"
	"time"
)

type TimedRateLimiter struct {
	// periodic forgetting of identifiers that have been seen & assigned a rate limiter to prevent bloat over time
	timers map[string]*time.Timer
	// buckets of access tokens, refreshing over time
	limiters map[string]*rate.Limiter
	// routes that are rate limited
	routes         map[string]bool
	limitAllRoutes bool
	refreshPeriod  time.Duration
	timeToRemember time.Duration
	burst          int
}

func NewTimedRateLimiter(limitedRoutes []string, refresh, remember time.Duration) *TimedRateLimiter {
	rl := TimedRateLimiter{}
	rl.timers = make(map[string]*time.Timer)
	rl.limiters = make(map[string]*rate.Limiter)
	rl.routes = make(map[string]bool)
	for _, route := range limitedRoutes {
		rl.routes[route] = true
	}
	rl.refreshPeriod = refresh
	rl.timeToRemember = remember
	rl.burst = 15 /* default value, use rl.SetBurstAllowance to change */
	return &rl
}

// amount of accesses allowed ~concurrently, before needing to wait for a rl.refreshPeriod
func (rl *TimedRateLimiter) SetBurstAllowance(burst int) {
	if burst >= 1 {
		rl.burst = burst
	}
}

func (rl *TimedRateLimiter) SetLimitAllRoutes(limitAll bool) {
	rl.limitAllRoutes = limitAll
}

// find out if resource access is allowed or not: calling consumes a rate limit token
func (rl *TimedRateLimiter) IsLimited(identifier, route string) bool {
	if !rl.limitAllRoutes {
		// route isn't rate limited
		if _, exists := rl.routes[route]; !exists {
			return false
		}
	}
	// route is designated to be rate limited, try the limiter to see if we can access it
	ret := !rl.access(identifier)
	return ret
}

func (rl *TimedRateLimiter) BlockUntilAllowed(identifier, route string, ctx context.Context) error {
	// route isn't rate limited
	if !rl.limitAllRoutes {
		if _, exists := rl.routes[route]; !exists {
			return nil
		}
	}
	limiter := rl.getLimiter(identifier)
	err := limiter.Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (rl *TimedRateLimiter) getLimiter(identifier string) *rate.Limiter {
	// limiter doesn't yet exist for this identifier
	if _, exists := rl.limiters[identifier]; !exists {
		// create a rate limit for it
		rl.createRateLimit(identifier)
		// remember this identifier (remote ip) for rl.timeToRemember before forgetting
		rl.rememberIdentifier(identifier)
	}
	limiter := rl.limiters[identifier]
	return limiter
}

// returns true if identifier currently allowed to access the resource
func (rl *TimedRateLimiter) access(identifier string) bool {
	limiter := rl.getLimiter(identifier)
	// consumes one token from the rate limiter bucket
	allowed := limiter.Allow()
	return allowed
}

func (rl *TimedRateLimiter) createRateLimit(identifier string) {
	accessRate := rate.Every(rl.refreshPeriod)
	limit := rate.NewLimiter(accessRate, rl.burst)
	rl.limiters[identifier] = limit
}

func (rl *TimedRateLimiter) rememberIdentifier(identifier string) {
	// timer already exists; refresh it
	if timer, exists := rl.timers[identifier]; exists {
		timer.Reset(rl.timeToRemember)
		return
	}
	// new timer
	timer := time.AfterFunc(rl.timeToRemember, func() {
		rl.forgetLimiter(identifier)
	})
	// map timer to its identifier
	rl.timers[identifier] = timer
}

// forget the rate limiter associated for this identifier (to prevent memory growth over time)
func (rl *TimedRateLimiter) forgetLimiter(identifier string) {
	if _, exists := rl.limiters[identifier]; exists {
		delete(rl.limiters, identifier)
	}
}
