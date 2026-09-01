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
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	auth2 "code.vikunja.io/api/pkg/modules/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

const rateLimitTestTimeout = 5 * time.Second

func receiveRateLimitTestValue[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(rateLimitTestTimeout):
		t.Fatal("timed out waiting for rate limit test")
		var zero T
		return zero
	}
}

type contextCheckingRateLimitStore struct {
	limiter.Store
	refunds atomic.Int32
}

func (s *contextCheckingRateLimitStore) Increment(ctx context.Context, key string, count int64, rate limiter.Rate) (limiter.Context, error) {
	if err := ctx.Err(); err != nil {
		return limiter.Context{}, err
	}
	if count < 0 {
		s.refunds.Add(1)
	}
	return s.Store.Increment(ctx, key, count, rate)
}

func newRateLimitTestContext(e *echo.Echo, remoteAddr string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, "/api/v2/info", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func rateLimitTestHandler(rateLimitKind string, limit int64) (handler echo.HandlerFunc, called *int) {
	calls := 0
	rateLimiter := limiter.New(memory.NewStore(), limiter.Rate{Period: time.Minute, Limit: limit})
	h := RateLimit(rateLimiter, rateLimitKind)(func(_ *echo.Context) error {
		calls++
		return nil
	})
	return h, &calls
}

// TestRateLimitUnauthenticated makes sure an unauthenticated request does not
// panic when the rate limit kind is "user" - the v2 group rate limits its public
// routes as well, so there are no jwt claims to key on.
func TestRateLimitUnauthenticated(t *testing.T) {
	log.InitLogger()
	e := echo.New()

	t.Run("does not panic and passes through", func(t *testing.T) {
		h, calls := rateLimitTestHandler("user", 100)
		c, rec := newRateLimitTestContext(e, "1.2.3.4:1234")

		require.NotPanics(t, func() {
			require.NoError(t, h(c))
		})
		assert.Equal(t, 1, *calls)
		assert.Equal(t, "100", rec.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "99", rec.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("limits by ip", func(t *testing.T) {
		h, calls := rateLimitTestHandler("user", 2)

		for range 2 {
			c, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
			require.NoError(t, h(c))
		}

		c, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
		err := h(c)
		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		assert.Equal(t, http.StatusTooManyRequests, he.Code)

		// A different ip must have its own bucket, otherwise everyone shares one limit
		other, _ := newRateLimitTestContext(e, "5.6.7.8:1234")
		require.NoError(t, h(other))
		assert.Equal(t, 3, *calls)
	})
}

// TestRateLimitUser makes sure authenticated requests are still keyed by user id.
func TestRateLimitUser(t *testing.T) {
	log.InitLogger()
	e := echo.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type":     float64(auth2.AuthTypeUser),
		"id":       float64(42),
		"username": "user42",
	})

	h, calls := rateLimitTestHandler("user", 2)

	for range 2 {
		c, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
		c.Set("user", token)
		require.NoError(t, h(c))
	}

	// Same user from a different ip shares the bucket
	c, _ := newRateLimitTestContext(e, "5.6.7.8:1234")
	c.Set("user", token)
	err := h(c)
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusTooManyRequests, he.Code)
	assert.Equal(t, 2, *calls)
}

// TestRateLimitUnknownKind makes sure a misconfigured kind falls back to per-ip
// limiting instead of putting every request into one shared bucket.
func TestRateLimitUnknownKind(t *testing.T) {
	log.InitLogger()
	e := echo.New()

	h, calls := rateLimitTestHandler("something-invalid", 1)

	c, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
	require.NoError(t, h(c))

	c, _ = newRateLimitTestContext(e, "1.2.3.4:1234")
	err := h(c)
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusTooManyRequests, he.Code)

	other, _ := newRateLimitTestContext(e, "5.6.7.8:1234")
	require.NoError(t, h(other))
	assert.Equal(t, 2, *calls)
}

func TestBasicAuthRateLimitBoundsConcurrentAuthentication(t *testing.T) {
	log.InitLogger()
	previousStore := config.RateLimitStore.GetString()
	previousLimit := config.RateLimitBasicAuthLimit.GetString()
	config.RateLimitStore.Set("memory")
	config.RateLimitBasicAuthLimit.Set("2")
	t.Cleanup(func() {
		config.RateLimitStore.Set(previousStore)
		config.RateLimitBasicAuthLimit.Set(previousLimit)
	})

	e := echo.New()
	entered := make(chan struct{}, 3)
	release := make(chan struct{})
	completed := make(chan int, 3)
	var calls atomic.Int32
	h := basicAuthRateLimit()(func(_ *echo.Context) error {
		calls.Add(1)
		entered <- struct{}{}
		<-release
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	})

	start := make(chan struct{})
	for range 3 {
		go func() {
			<-start
			c, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
			c.Request().SetBasicAuth("user", "wrong-password")
			completed <- echo.StatusCode(h(c))
		}()
	}
	close(start)

	receiveRateLimitTestValue(t, entered)
	receiveRateLimitTestValue(t, entered)
	thirdStatus := 0
	completedCount := 0
	select {
	case <-entered:
	case thirdStatus = <-completed:
		completedCount++
	case <-time.After(rateLimitTestTimeout):
		t.Fatal("timed out waiting for the third authentication")
	}
	close(release)
	for completedCount < 3 {
		status := receiveRateLimitTestValue(t, completed)
		completedCount++
		if status == http.StatusTooManyRequests {
			thirdStatus = status
		}
	}

	assert.LessOrEqual(t, calls.Load(), int32(2))
	assert.Equal(t, http.StatusTooManyRequests, thirdStatus)
}

func TestBasicAuthRateLimitDoesNotRefundAnExpiredReservation(t *testing.T) {
	log.InitLogger()
	e := echo.New()
	store := &contextCheckingRateLimitStore{Store: memory.NewStore()}
	rateLimiter := limiter.New(store, limiter.Rate{Period: 1100 * time.Millisecond, Limit: 1})
	initialNow := time.Now()
	var nowNanos atomic.Int64
	nowNanos.Store(initialNow.UnixNano())
	now := func() time.Time {
		return time.Unix(0, nowNanos.Load())
	}

	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstCompleted := make(chan error, 1)
	var calls atomic.Int32
	h := basicAuthRateLimitWithClock(rateLimiter, now)(func(_ *echo.Context) error {
		if calls.Add(1) == 1 {
			close(firstEntered)
			<-releaseFirst
			return nil
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	})

	go func() {
		c, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
		c.Request().SetBasicAuth("user", "password")
		firstCompleted <- h(c)
	}()
	receiveRateLimitTestValue(t, firstEntered)

	nowNanos.Store(initialNow.Add(time.Minute).UnixNano())

	second, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
	second.Request().SetBasicAuth("user", "wrong-password")
	assert.Equal(t, http.StatusUnauthorized, echo.StatusCode(h(second)))

	close(releaseFirst)
	require.NoError(t, receiveRateLimitTestValue(t, firstCompleted))
	assert.Zero(t, store.refunds.Load())

	third, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
	third.Request().SetBasicAuth("user", "wrong-password")
	assert.Equal(t, http.StatusTooManyRequests, echo.StatusCode(h(third)))
	assert.Equal(t, int32(2), calls.Load())
}

func TestBasicAuthRateLimitRefundIgnoresCanceledRequestContext(t *testing.T) {
	log.InitLogger()
	e := echo.New()
	store := &contextCheckingRateLimitStore{Store: memory.NewStore()}
	rateLimiter := limiter.New(store, limiter.Rate{Period: time.Minute, Limit: 1})

	var cancel context.CancelFunc
	h := basicAuthRateLimitWithLimiter(rateLimiter)(func(_ *echo.Context) error {
		cancel()
		return nil
	})

	c, _ := newRateLimitTestContext(e, "1.2.3.4:1234")
	requestContext, requestCancel := context.WithCancel(c.Request().Context())
	c.SetRequest(c.Request().WithContext(requestContext))
	c.Request().SetBasicAuth("user", "password")
	cancel = requestCancel
	t.Cleanup(requestCancel)

	require.NoError(t, h(c))

	assert.Equal(t, int32(1), store.refunds.Load())
}
