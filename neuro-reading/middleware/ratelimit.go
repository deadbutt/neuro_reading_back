package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func init() {
	go cleanupVisitors()
}

func cleanupVisitors() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{lastSeen: time.Now(), count: 1}
			mu.Unlock()
			c.Next()
			return
		}

		if time.Since(v.lastSeen) > time.Minute {
			v.count = 1
			v.lastSeen = time.Now()
		} else {
			v.count++
		}
		mu.Unlock()

		if v.count > 100 {
			c.JSON(429, gin.H{"code": 1001, "message": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}

		c.Next()
	}
}
