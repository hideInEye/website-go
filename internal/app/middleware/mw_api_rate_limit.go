package middleware

import (
	"strconv"
	"sync"
	"time"

	"go-gin/internal/app/ginplus"

	"github.com/gin-gonic/gin"
)

// APIRateLimit applies the original per-handler, per-IP throttling without Redis.
func APIRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	type bucket struct {
		count int
		until time.Time
	}

	var mutex sync.Mutex
	buckets := make(map[string]bucket)
	nextCleanup := time.Now().Add(window)
	return func(c *gin.Context) {
		now := time.Now()
		key := c.FullPath() + ":" + c.ClientIP()
		mutex.Lock()
		if !now.Before(nextCleanup) || len(buckets) >= 16384 {
			for key, value := range buckets {
				if !now.Before(value.until) {
					delete(buckets, key)
				}
			}
			nextCleanup = now.Add(window)
		}
		value, exists := buckets[key]
		if !exists && len(buckets) >= 16384 {
			mutex.Unlock()
			c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
			ginplus.Fail(c, 429, "请求过于频繁，请稍后再试")
			return
		}
		if !exists || !now.Before(value.until) {
			value = bucket{until: now.Add(window)}
		}
		value.count++
		buckets[key] = value
		mutex.Unlock()
		remaining := limit - value.count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(value.until.Unix(), 10))
		if value.count > limit {
			c.Header("Retry-After", strconv.Itoa(int(value.until.Sub(now).Seconds())+1))
			ginplus.Fail(c, 429, "请求过于频繁，请稍后再试")
			return
		}
		c.Next()
	}
}
