package middleware

import (
	"context"
	"fmt"
	"neuro-reading/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		ctx := context.Background()

		now := time.Now().Unix()
		window := int64(60)
		limit := int64(100)

		pipe := db.Redis.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now-window, 10))
		pipe.ZCard(ctx, key)
		results, err := pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		count := results[1].(*redis.IntCmd).Val()
		if count >= limit {
			c.JSON(429, gin.H{"code": 1001, "message": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}

		db.Redis.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: now,
		})
		db.Redis.Expire(ctx, key, time.Duration(window)*time.Second)

		c.Next()
	}
}
