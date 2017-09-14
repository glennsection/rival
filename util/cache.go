package util

import (
	"net/url"

	"github.com/garyburd/redigo/redis"

	"bloodtales/log"
)

type Cache struct {
	Stream

	// internal
	redis    redis.Conn
}

type CacheStreamSource struct {
	// internal
	redis    redis.Conn
}

var (
	// internal
	redisPool      *redis.Pool
	redisURL       url.URL
	redisPassword  string = ""
)

func (source CacheStreamSource) Has(name string) bool {
	ok, _ := redis.Bool(source.redis.Do("EXISTS", name))
	return ok
}

func (source CacheStreamSource) Set(name string, value interface{}) {
	var err error
	if IsNil(value) {
		_, err = source.redis.Do("DEL", name)
	} else {
		_, err = source.redis.Do("SET", name, value)
	}
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
}

func (source CacheStreamSource) Get(name string) interface{} {
	value, err := redis.String(source.redis.Do("GET", name))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorf("Redis error: %v", err)
			PrintStack()
		}
		return ""
	}
	return value
}

func init() {
	// get redis URL
	rawRedisURL := Env.GetRequiredString("REDIS_URL")
	redisURL, err := url.Parse(rawRedisURL)
	if err != nil {
		return
	}

	// get password
	if redisURL.User != nil {
		if password, ok := redisURL.User.Password(); ok {
			redisPassword = password
		}
	}

	// TODO - select DB?
	// if len(redisURL.Path) > 1 {
	// 	db := strings.TrimPrefix(redisURL.Path, "/")
	// 	c.Do("SELECT", db)
	// }

	// connect and create redis pool
	redisPool = &redis.Pool {
		// Maximum number of idle connections in the pool.
		MaxIdle: Env.GetInt("REDIS_MAX_IDLE", 80),

		// Maximum number of connections allocated by the pool at a given time.
		// When zero, there is no limit on the number of connections in the pool.
		MaxActive: Env.GetInt("REDIS_MAX_ACTIVE", 12000),

		// function to create new connection
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisURL.Host)
			if err != nil {
				return nil, err
			}

			// password authentication
			if redisPassword != "" {
				if _, err := c.Do("AUTH", redisPassword); err != nil {
					c.Close()
					return nil, err
				}
			}

			// TODO - select DB?
			// if _, err := c.Do("SELECT", db); err != nil {
			// 	c.Close()
			// 	return nil, err
			// }

			return c, nil
		},
	}
}

func CloseCache() {
	// cleanup redis pool
	if redisPool != nil {
		redisPool.Close()
	}
}

func GetCacheConnection() (cache *Cache) {
	// get redis connection from pool
	redis := redisPool.Get()

	// stream source
	source := CacheStreamSource {
		redis: redis,
	}

	// create abstracted cache
	cache = &Cache {
		Stream: Stream {
			source: source,
		},
		redis: redis,
	}
	return
}

func (cache *Cache) Close() {
	// close redis connection
	cache.redis.Close()
}

func (cache *Cache) Expire(name string, ttl int) {
	var err error
	_, err = cache.redis.Do("EXPIRE", name, ttl)
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
}

func (cache *Cache) SetScore(group string, name string, score int) {
	_, err := cache.redis.Do("ZADD", group, score, name)
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
}

func (cache *Cache) GetScore(group string, name string) int {
	result, err := redis.Int(cache.redis.Do("ZSCORE", group, name))
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
	return result
}

func (cache *Cache) RemoveScore(group string, name string) {
	_, err := cache.redis.Do("ZREM", group, name)
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
}

func (cache *Cache) ClearScores(group string) {
	_, err := cache.redis.Do("DEL", group)
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
}

func (cache *Cache) GetRank(group string, name string) int {
	result, err := redis.Int(cache.redis.Do("ZRANK", group, name))
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
	return result
}

func (cache *Cache) GetRankRange(group string, start int, stop int) []string {
	result, err := redis.Strings(cache.redis.Do("ZREVRANGE", group, start, stop))
	if err != nil {
		log.Errorf("Redis error: %v", err)
		PrintStack()
	}
	return result
}
