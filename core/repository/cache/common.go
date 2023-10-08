package cache

import (
	"github.com/jellydator/ttlcache/v3"
	"time"
)

type Cache struct {
	cache *ttlcache.Cache[string, string]
}

func NewCacheCommon() *Cache {
	return &Cache{
		cache: ttlcache.New[string, string](
			ttlcache.WithTTL[string, string](59 * time.Minute),
		),
	}
}

func (c *Cache) Set(key, token string, duration time.Duration) {
	if duration == 0 {
		duration = 5 * time.Minute
	}

	c.cache.Set(key, token, duration)
}

func (c *Cache) Get(key string) string {
	item := c.cache.Get(key)
	if item == nil {
		return ""
	}

	if item.ExpiresAt().Before(time.Now()) {
		c.cache.Delete(key)
		return ""
	}

	return item.Value()
}
