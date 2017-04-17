package system

import (
	"net/url"

	"github.com/garyburd/redigo/redis"
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

func (source CacheStreamSource) Set(name string, value interface{}) {
	_, err := source.redis.Do("SET", name, value)
	if err != nil {
		panic(err)
	}
}

func (source CacheStreamSource) Get(name string) interface{} {
	value, err := redis.String(source.redis.Do("GET", name))
	if err == nil {
		return value
	}
	return ""
}

func (application *Application) initializeCache() {
	// get redis URL
	rawRedisURL := application.Env.GetRequiredString("REDIS_URL")
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
		MaxIdle:   application.Env.GetInt("REDIS_MAX_IDLE", 80),

		// Maximum number of connections allocated by the pool at a given time.
		// When zero, there is no limit on the number of connections in the pool.
		MaxActive: application.Env.GetInt("REDIS_MAX_ACTIVE", 12000),

		// function to create new connection
		Dial:      func() (redis.Conn, error) {
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

func (application *Application) closeCache() {
	// cleanup redis pool
	if redisPool != nil {
		redisPool.Close()
	}
}

func (application *Application) GetCache() (cache *Cache) {
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
