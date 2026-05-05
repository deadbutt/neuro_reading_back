package middleware

import (
	"neuro-reading/config"
	"neuro-reading/model"
	"neuro-reading/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, model.Error(1002, "未授权/Token过期"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, model.Error(1002, "未授权/Token过期"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1], cfg)
		if err != nil {
			c.JSON(401, model.Error(1002, "未授权/Token过期"))
			c.Abort()
			return
		}

		if claims.Type != "access" {
			c.JSON(401, model.Error(1002, "未授权/Token过期"))
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Next()
	}
}