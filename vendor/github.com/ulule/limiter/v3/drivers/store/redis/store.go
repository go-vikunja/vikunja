package redis

import (
	"context"
	"fmt"
	"time"

	libredis "github.com/go-redis/redis/v7"
	"github.com/pkg/errors"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/common"
)

// Client is an interface thats allows to use a redis cluster or a redis single client seamlessly.
type Client interface {
	Ping() *libredis.StatusCmd
	Get(key string) *libredis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *libredis.StatusCmd
	Watch(handler func(*libredis.Tx) error, keys ...string) error
	Del(keys ...string) *libredis.IntCmd
	SetNX(key string, value interface{}, expiration time.Duration) *libredis.BoolCmd
	Eval(script string, keys []string, args ...interface{}) *libredis.Cmd
}

// Store is the redis store.
type Store struct {
	// Prefix used for the key.
	Prefix string
	// MaxRetry is the maximum number of retry under race conditions.
	MaxRetry int
	// client used to communicate with redis server.
	client Client
}

// NewStore returns an instance of redis store with defaults.
func NewStore(client Client) (limiter.Store, error) {
	return NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:          limiter.DefaultPrefix,
		CleanUpInterval: limiter.DefaultCleanUpInterval,
		MaxRetry:        limiter.DefaultMaxRetry,
	})
}

// NewStoreWithOptions returns an instance of redis store with options.
func NewStoreWithOptions(client Client, options limiter.StoreOptions) (limiter.Store, error) {
	store := &Store{
		client:   client,
		Prefix:   options.Prefix,
		MaxRetry: options.MaxRetry,
	}

	if store.MaxRetry <= 0 {
		store.MaxRetry = 1
	}

	_, err := store.ping()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Get returns the limit for given identifier.
func (store *Store) Get(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	now := time.Now()

	lctx := limiter.Context{}
	onWatch := func(rtx *libredis.Tx) error {

		created, err := store.doSetValue(rtx, key, rate.Period)
		if err != nil {
			return err
		}

		if created {
			expiration := now.Add(rate.Period)
			lctx = common.GetContextFromState(now, rate, expiration, 1)
			return nil
		}

		count, ttl, err := store.doUpdateValue(rtx, key, rate.Period)
		if err != nil {
			return err
		}

		expiration := now.Add(rate.Period)
		if ttl > 0 {
			expiration = now.Add(ttl)
		}

		lctx = common.GetContextFromState(now, rate, expiration, count)
		return nil
	}

	err := store.client.Watch(onWatch, key)
	if err != nil {
		err = errors.Wrapf(err, "limiter: cannot get value for %s", key)
		return limiter.Context{}, err
	}

	return lctx, nil
}

// Peek returns the limit for given identifier, without modification on current values.
func (store *Store) Peek(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	now := time.Now()

	lctx := limiter.Context{}
	onWatch := func(rtx *libredis.Tx) error {
		count, ttl, err := store.doPeekValue(rtx, key)
		if err != nil {
			return err
		}

		expiration := now.Add(rate.Period)
		if ttl > 0 {
			expiration = now.Add(ttl)
		}

		lctx = common.GetContextFromState(now, rate, expiration, count)
		return nil
	}

	err := store.client.Watch(onWatch, key)
	if err != nil {
		err = errors.Wrapf(err, "limiter: cannot peek value for %s", key)
		return limiter.Context{}, err
	}

	return lctx, nil
}

// Reset returns the limit for given identifier which is set to zero.
func (store *Store) Reset(ctx context.Context, key string, rate limiter.Rate) (limiter.Context, error) {
	key = fmt.Sprintf("%s:%s", store.Prefix, key)
	now := time.Now()

	lctx := limiter.Context{}
	onWatch := func(rtx *libredis.Tx) error {

		err := store.doResetValue(rtx, key)
		if err != nil {
			return err
		}

		count := int64(0)
		expiration := now.Add(rate.Period)

		lctx = common.GetContextFromState(now, rate, expiration, count)
		return nil
	}

	err := store.client.Watch(onWatch, key)
	if err != nil {
		err = errors.Wrapf(err, "limiter: cannot reset value for %s", key)
		return limiter.Context{}, err
	}

	return lctx, nil
}

// doPeekValue will execute peekValue with a retry mecanism (optimistic locking) until store.MaxRetry is reached.
func (store *Store) doPeekValue(rtx *libredis.Tx, key string) (int64, time.Duration, error) {
	for i := 0; i < store.MaxRetry; i++ {
		count, ttl, err := peekValue(rtx, key)
		if err == nil {
			return count, ttl, nil
		}
	}
	return 0, 0, errors.New("retry limit exceeded")
}

// peekValue will retrieve the counter and its expiration for given key.
func peekValue(rtx *libredis.Tx, key string) (int64, time.Duration, error) {
	pipe := rtx.TxPipeline()
	value := pipe.Get(key)
	expire := pipe.PTTL(key)

	_, err := pipe.Exec()
	if err != nil && err != libredis.Nil {
		return 0, 0, err
	}

	count, err := value.Int64()
	if err != nil && err != libredis.Nil {
		return 0, 0, err
	}

	ttl, err := expire.Result()
	if err != nil {
		return 0, 0, err
	}

	return count, ttl, nil
}

// doSetValue will execute setValue with a retry mecanism (optimistic locking) until store.MaxRetry is reached.
func (store *Store) doSetValue(rtx *libredis.Tx, key string, expiration time.Duration) (bool, error) {
	for i := 0; i < store.MaxRetry; i++ {
		created, err := setValue(rtx, key, expiration)
		if err == nil {
			return created, nil
		}
	}
	return false, errors.New("retry limit exceeded")
}

// setValue will try to initialize a new counter if given key doesn't exists.
func setValue(rtx *libredis.Tx, key string, expiration time.Duration) (bool, error) {
	value := rtx.SetNX(key, 1, expiration)

	created, err := value.Result()
	if err != nil {
		return false, err
	}

	return created, nil
}

// doUpdateValue will execute setValue with a retry mecanism (optimistic locking) until store.MaxRetry is reached.
func (store *Store) doUpdateValue(rtx *libredis.Tx, key string,
	expiration time.Duration) (int64, time.Duration, error) {
	for i := 0; i < store.MaxRetry; i++ {
		count, ttl, err := updateValue(rtx, key, expiration)
		if err == nil {
			return count, ttl, nil
		}

		// If ttl is negative and there is an error, do not retry an update.
		if ttl < 0 {
			return 0, 0, err
		}
	}
	return 0, 0, errors.New("retry limit exceeded")
}

// updateValue will try to increment the counter identified by given key.
func updateValue(rtx *libredis.Tx, key string, expiration time.Duration) (int64, time.Duration, error) {
	pipe := rtx.TxPipeline()
	value := pipe.Incr(key)
	expire := pipe.PTTL(key)

	_, err := pipe.Exec()
	if err != nil {
		return 0, 0, err
	}

	count, err := value.Result()
	if err != nil {
		return 0, 0, err
	}

	ttl, err := expire.Result()
	if err != nil {
		return 0, 0, err
	}

	// If ttl is less than zero, we have to define key expiration.
	// The PTTL command returns -2 if the key does not exist, and -1 if the key exists, but there is no expiry set.
	// We shouldn't try to set an expiry on a key that doesn't exist.
	if isExpirationRequired(ttl) {
		expire := rtx.Expire(key, expiration)

		ok, err := expire.Result()
		if err != nil {
			return count, ttl, err
		}

		if !ok {
			return count, ttl, errors.New("cannot configure timeout on key")
		}
	}

	return count, ttl, nil
}

// doResetValue will execute resetValue with a retry mecanism (optimistic locking) until store.MaxRetry is reached.
func (store *Store) doResetValue(rtx *libredis.Tx, key string) error {
	for i := 0; i < store.MaxRetry; i++ {
		err := resetValue(rtx, key)
		if err == nil {
			return nil
		}
	}
	return errors.New("retry limit exceeded")
}

// resetValue will try to reset the counter identified by given key.
func resetValue(rtx *libredis.Tx, key string) error {
	deletion := rtx.Del(key)

	_, err := deletion.Result()
	if err != nil {
		return err
	}

	return nil
}

// ping checks if redis is alive.
func (store *Store) ping() (bool, error) {
	cmd := store.client.Ping()

	pong, err := cmd.Result()
	if err != nil {
		return false, errors.Wrap(err, "limiter: cannot ping redis server")
	}

	return (pong == "PONG"), nil
}

// isExpirationRequired returns if we should set an expiration on a key, using (error) result from PTTL command.
// The error code is -2 if the key does not exist, and -1 if the key exists.
// Usually, it should be returned in nanosecond, but some users have reported that it could be in millisecond as well.
// Better safe than sorry: we handle both.
func isExpirationRequired(ttl time.Duration) bool {
	switch ttl {
	case -1 * time.Nanosecond, -1 * time.Millisecond:
		return true
	default:
		return false
	}
}
