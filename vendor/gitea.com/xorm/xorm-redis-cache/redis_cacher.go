package xormrediscache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"reflect"
	"time"
	"unsafe"

	"github.com/garyburd/redigo/redis"
	"xorm.io/xorm/caches"
	"xorm.io/xorm/log"
)

const (
	DEFAULT_EXPIRATION = time.Duration(0)
	FOREVER_EXPIRATION = time.Duration(-1)

	LOGGING_PREFIX = "[redis_cacher]"
)

// RedisCacher wraps the Redis client to meet the Cache interface.
type RedisCacher struct {
	pool              *redis.Pool
	defaultExpiration time.Duration

	Logger log.ContextLogger
}

// NewRedisCacher creates a Redis Cacher, host as IP endpoint, i.e., localhost:6379, provide empty string or nil if Redis server doesn't
// require AUTH command, defaultExpiration sets the expire duration for a key to live. Until redigo supports
// sharding/clustering, only one host will be in hostList
//
//     engine.SetDefaultCacher(xormrediscache.NewRedisCacher("localhost:6379", "", xormrediscache.DEFAULT_EXPIRATION, engine.Logger))
//
// or set MapCacher
//
//     engine.MapCacher(&user, xormrediscache.NewRedisCacher("localhost:6379", "", xormrediscache.DEFAULT_EXPIRATION, engine.Logger))
//
func NewRedisCacher(host string, password string, defaultExpiration time.Duration, logger log.ContextLogger) *RedisCacher {
	var pool = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			// the redis protocol should probably be made sett-able
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			} else {
				// check with PING
				if _, err := c.Do("PING"); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		// custom connection test method
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if _, err := c.Do("PING"); err != nil {
				return err
			}
			return nil
		},
	}
	return MakeRedisCacher(pool, defaultExpiration, logger)
}

// MakeRedisCacher build a cacher based on redis.Pool
func MakeRedisCacher(pool *redis.Pool, defaultExpiration time.Duration, logger log.ContextLogger) *RedisCacher {
	return &RedisCacher{pool: pool, defaultExpiration: defaultExpiration, Logger: logger}
}

func exists(conn redis.Conn, key string) bool {
	existed, _ := redis.Bool(conn.Do("EXISTS", key))
	return existed
}

func (c *RedisCacher) logErrf(format string, contents ...interface{}) {
	if c.Logger != nil {
		c.Logger.Errorf(fmt.Sprintf("%s %s", LOGGING_PREFIX, format), contents...)
	}
}

func (c *RedisCacher) logDebugf(format string, contents ...interface{}) {
	if c.Logger != nil {
		c.Logger.Debugf(fmt.Sprintf("%s %s", LOGGING_PREFIX, format), contents...)
	}
}

func (c *RedisCacher) getBeanKey(tableName string, id string) string {
	return fmt.Sprintf("xorm:bean:%s:%s", tableName, id)
}

func (c *RedisCacher) getSqlKey(tableName string, sql string) string {
	// hash sql to minimize key length
	crc := crc32.ChecksumIEEE([]byte(sql))
	return fmt.Sprintf("xorm:sql:%s:%d", tableName, crc)
}

// Flush deletes all xorm cached objects
func (c *RedisCacher) Flush() error {
	// conn := c.pool.Get()
	// defer conn.Close()
	// _, err := conn.Do("FLUSHALL")
	// return err
	return c.delObject("xorm:*")
}

func (c *RedisCacher) getObject(key string) interface{} {
	conn := c.pool.Get()
	defer conn.Close()
	raw, err := conn.Do("GET", key)
	if raw == nil {
		return nil
	}
	item, err := redis.Bytes(raw, err)
	if err != nil {
		c.logErrf("redis.Bytes failed: %s", err)
		return nil
	}

	value, err := c.deserialize(item)

	return value
}

func (c *RedisCacher) GetIds(tableName, sql string) interface{} {
	sqlKey := c.getSqlKey(tableName, sql)
	c.logDebugf(" GetIds|tableName:%s|sql:%s|key:%s", tableName, sql, sqlKey)
	return c.getObject(sqlKey)
}

func (c *RedisCacher) GetBean(tableName string, id string) interface{} {
	beanKey := c.getBeanKey(tableName, id)
	c.logDebugf("[xorm/redis_cacher] GetBean|tableName:%s|id:%s|key:%s", tableName, id, beanKey)
	return c.getObject(beanKey)
}

func (c *RedisCacher) putObject(key string, value interface{}) {
	c.invoke(c.pool.Get().Do, key, value, c.defaultExpiration)
}

func (c *RedisCacher) PutIds(tableName, sql string, ids interface{}) {
	sqlKey := c.getSqlKey(tableName, sql)
	c.logDebugf("PutIds|tableName:%s|sql:%s|key:%s|obj:%s|type:%v", tableName, sql, sqlKey, ids, reflect.TypeOf(ids))
	c.putObject(sqlKey, ids)
}

func (c *RedisCacher) PutBean(tableName string, id string, obj interface{}) {
	beanKey := c.getBeanKey(tableName, id)
	c.logDebugf("PutBean|tableName:%s|id:%s|key:%s|type:%v", tableName, id, beanKey, reflect.TypeOf(obj))
	c.putObject(beanKey, obj)
}

func (c *RedisCacher) delObject(key string) error {
	c.logDebugf("delObject key:[%s]", key)

	conn := c.pool.Get()
	defer conn.Close()
	if !exists(conn, key) {
		c.logErrf("delObject key:[%s] err: %v", key, caches.ErrCacheMiss)
		return caches.ErrCacheMiss
	}
	_, err := conn.Do("DEL", key)
	return err
}

func (c *RedisCacher) delObjects(key string) error {

	c.logDebugf("delObjects key:[%s]", key)

	conn := c.pool.Get()
	defer conn.Close()

	keys, err := conn.Do("KEYS", key)
	c.logDebugf("delObjects keys: %v", keys)

	if err == nil {
		for _, key := range keys.([]interface{}) {
			conn.Do("DEL", key)
		}
	}
	return err
}

func (c *RedisCacher) DelIds(tableName, sql string) {
	c.delObject(c.getSqlKey(tableName, sql))
}

func (c *RedisCacher) DelBean(tableName string, id string) {
	c.delObject(c.getBeanKey(tableName, id))
}

func (c *RedisCacher) ClearIds(tableName string) {
	c.delObjects(fmt.Sprintf("xorm:sql:%s:*", tableName))
}

func (c *RedisCacher) ClearBeans(tableName string) {
	c.delObjects(c.getBeanKey(tableName, "*"))
}

func (c *RedisCacher) invoke(f func(string, ...interface{}) (interface{}, error),
	key string, value interface{}, expires time.Duration) error {

	switch expires {
	case DEFAULT_EXPIRATION:
		expires = c.defaultExpiration
	case FOREVER_EXPIRATION:
		expires = time.Duration(0)
	}

	b, err := c.serialize(value)
	if err != nil {
		return err
	}
	conn := c.pool.Get()
	defer conn.Close()
	if expires > 0 {
		_, err := f("SETEX", key, int32(expires/time.Second), b)
		return err
	} else {
		_, err := f("SET", key, b)
		return err
	}
}

func (c *RedisCacher) serialize(value interface{}) ([]byte, error) {

	err := c.registerGobConcreteType(value)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(value).Kind() == reflect.Struct {
		return nil, fmt.Errorf("serialize func only take pointer of a struct")
	}

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)

	c.logDebugf("serialize type:%v", reflect.TypeOf(value))
	err = encoder.Encode(&value)
	if err != nil {
		c.logErrf("gob encoding '%s' failed: %s|value:%v", value, err, value)
		return nil, err
	}
	return b.Bytes(), nil
}

func (c *RedisCacher) deserialize(byt []byte) (ptr interface{}, err error) {
	b := bytes.NewBuffer(byt)
	decoder := gob.NewDecoder(b)

	var p interface{}
	err = decoder.Decode(&p)
	if err != nil {
		c.logErrf("decode failed: %v", err)
		return
	}

	v := reflect.ValueOf(p)
	c.logDebugf("deserialize type:%v", v.Type())
	if v.Kind() == reflect.Struct {

		var pp interface{} = &p
		datas := reflect.ValueOf(pp).Elem().InterfaceData()

		sp := reflect.NewAt(v.Type(),
			unsafe.Pointer(datas[1])).Interface()
		ptr = sp
		vv := reflect.ValueOf(ptr)
		c.logDebugf("deserialize convert ptr type:%v | CanAddr:%t", vv.Type(), vv.CanAddr())
	} else {
		ptr = p
	}
	return
}

func (c *RedisCacher) registerGobConcreteType(value interface{}) error {

	t := reflect.TypeOf(value)

	c.logDebugf("registerGobConcreteType:%v", t)

	switch t.Kind() {
	case reflect.Ptr:
		v := reflect.ValueOf(value)
		i := v.Elem().Interface()
		gob.Register(&i)
	case reflect.Struct, reflect.Map, reflect.Slice:
		gob.Register(value)
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		// do nothing since already registered known type
	default:
		return fmt.Errorf("unhandled type: %v", t)
	}
	return nil
}

func (c *RedisCacher) GetPool() (*redis.Pool, error) {
	return c.pool, nil
}

func (c *RedisCacher) SetPool(pool *redis.Pool) {
	c.pool = pool
}
