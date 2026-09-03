// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package routes

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/red"
	"github.com/labstack/echo/v5"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"github.com/ulule/limiter/v3/drivers/store/redis"
)

// RateLimit is the rate limit middleware
func RateLimit(rateLimiter *limiter.Limiter, rateLimitKind string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			var rateLimitKey string
			switch rateLimitKind {
			case "ip":
				rateLimitKey = c.RealIP()
			case "user":
				// Unauthenticated requests hit this middleware because v2 rate limits
				// one group covering its public routes as well - key those by IP.
				rateLimitKey = "ip_" + c.RealIP()
				if auth2.HasAuthInContext(c) {
					auth, err := auth2.GetAuthFromClaims(c)
					if err != nil || auth == nil {
						log.Errorf("Error getting auth from jwt claims: %v", err)
					} else {
						rateLimitKey = "user_" + strconv.FormatInt(auth.GetID(), 10)
					}
				}
			default:
				log.Errorf("Unknown rate limit kind configured: %s", rateLimitKind)
				rateLimitKey = "ip_" + c.RealIP()
			}
			limiterCtx, err := rateLimiter.Get(c.Request().Context(), rateLimitKey)
			if err != nil {
				log.Errorf("IPRateLimit - rateLimiter.Get - err: %v, %s on %s", err, rateLimitKey, c.Request().URL)
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
			}

			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
			h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

			if limiterCtx.Reached {
				log.Infof("Too Many Requests from %s on %s", rateLimitKey, c.Request().URL)
				return echo.NewHTTPError(http.StatusTooManyRequests, "Too Many Requests")
			}

			// log.Printf("%s request continue", c.RealIP())
			return next(c)
		}
	}
}

// createRateLimiter builds a limiter with its own counters. The prefix keeps
// those counters apart in a shared backend: the redis store namespaces keys with
// it, so without a distinct one every limiter would decrement the same budget.
func createRateLimiter(prefix string, rate limiter.Rate) *limiter.Limiter {
	options := limiter.StoreOptions{
		Prefix:          limiter.DefaultPrefix + ":" + prefix,
		CleanUpInterval: limiter.DefaultCleanUpInterval,
		MaxRetry:        limiter.DefaultMaxRetry,
	}

	var store limiter.Store
	var err error
	switch config.RateLimitStore.GetString() {
	case "memory":
		store = memory.NewStoreWithOptions(options)
	case "redis":
		if !config.RedisEnabled.GetBool() {
			log.Fatal("Redis is configured for rate limiting, but not enabled!")
		}
		store, err = redis.NewStoreWithOptions(red.GetRedis(), options)
		if err != nil {
			log.Fatalf("Error while creating rate limit redis store: %s", err)
		}
	default:
		log.Fatalf("Unknown Rate limit store \"%s\"", config.RateLimitStore.GetString())
	}
	return limiter.New(store, rate)
}

// perMinuteIPRateLimit ignores RateLimitEnabled on purpose: pre-auth routes need
// a floor even with the global limiter off, which is the default.
func perMinuteIPRateLimit(prefix string, limit int64) echo.MiddlewareFunc {
	rate := limiter.Rate{
		Period: 60 * time.Second,
		Limit:  limit,
	}
	return RateLimit(createRateLimiter(prefix, rate), "ip")
}

func unauthRateLimit() echo.MiddlewareFunc {
	return perMinuteIPRateLimit("noauth", config.RateLimitNoAuthRoutesLimit.GetInt64())
}

// Renewal is routine traffic - sharing the credential-guessing floor let it starve /login.
func tokenRefreshRateLimit() echo.MiddlewareFunc {
	return perMinuteIPRateLimit("tokenrefresh", config.RateLimitTokenRefreshLimit.GetInt64())
}

// Limit failed bcrypt checks without throttling successful CalDAV syncs (GHSA-m469-88xx-8rx2).
func basicAuthRateLimit() echo.MiddlewareFunc {
	rate := limiter.Rate{
		Period: 60 * time.Second,
		Limit:  config.RateLimitBasicAuthLimit.GetInt64(),
	}
	return basicAuthRateLimitWithLimiter(createRateLimiter("basicauth", rate))
}

func basicAuthRateLimitWithLimiter(rateLimiter *limiter.Limiter) echo.MiddlewareFunc {
	return basicAuthRateLimitWithClock(rateLimiter, time.Now)
}

func basicAuthRateLimitWithClock(rateLimiter *limiter.Limiter, now func() time.Time) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			// A request without credentials is the normal 401 challenge, not a guess.
			if _, _, ok := c.Request().BasicAuth(); !ok {
				return next(c)
			}

			key := basicAuthRateLimitKey(c.RealIP(), rateLimiter.Rate.Period, now())
			// Reserve before authentication so concurrent guesses cannot bypass the limit.
			limiterCtx, err := rateLimiter.Increment(c.Request().Context(), key, 1)
			if err != nil {
				log.Errorf("basicAuthRateLimit - rateLimiter.Increment - err: %v, %s on %s", err, key, c.Request().URL)
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
			}
			if limiterCtx.Reached {
				if ierr := refundBasicAuthReservation(rateLimiter, key, limiterCtx.Reset, now()); ierr != nil {
					log.Errorf("basicAuthRateLimit - rateLimiter.Increment refund - err: %v", ierr)
					return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(ierr)
				}
				log.Infof("Too many failed basic auth attempts from %s on %s", key, c.Request().URL)
				return echo.NewHTTPError(http.StatusTooManyRequests, "Too Many Requests")
			}

			err = next(c)

			// A 401 may be returned or already written by BasicAuth middleware.
			failed := echo.StatusCode(err) == http.StatusUnauthorized
			if !failed {
				if res, ok := c.Response().(*echo.Response); ok && res.Committed && res.Status == http.StatusUnauthorized {
					failed = true
				}
			}
			if !failed {
				if ierr := refundBasicAuthReservation(rateLimiter, key, limiterCtx.Reset, now()); ierr != nil {
					log.Errorf("basicAuthRateLimit - rateLimiter.Increment refund - err: %v", ierr)
					return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(ierr)
				}
			}

			return err
		}
	}
}

func refundBasicAuthReservation(rateLimiter *limiter.Limiter, key string, reset int64, now time.Time) error {
	if now.Unix() > reset {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := rateLimiter.Increment(ctx, key, -1)
	return err
}

func basicAuthRateLimitKey(ip string, period time.Duration, now time.Time) string {
	// A late refund must never decrement a newer window.
	return ip + ":" + strconv.FormatInt(now.UnixNano()/period.Nanoseconds(), 10)
}

func setupRateLimit(a *echo.Group, rateLimitKind string) {
	if config.RateLimitEnabled.GetBool() {
		rate := limiter.Rate{
			Period: config.RateLimitPeriod.GetDuration() * time.Second,
			Limit:  config.RateLimitLimit.GetInt64(),
		}
		rateLimiter := createRateLimiter("global", rate)
		log.Debugf("Rate limit configured with %s and %v requests per %v", config.RateLimitStore.GetString(), rate.Limit, rate.Period)
		a.Use(RateLimit(rateLimiter, rateLimitKind))
	}
}
