package middleware

import (
	"strings"

	"ticketguard/backend/domain"
	"ticketguard/backend/utils"

	"github.com/gin-gonic/gin"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.Error(c, utils.ErrUnauthorized)
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseJWT(tokenString, secret)
		if err != nil {
			utils.Error(c, err)
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get("role")
		if !exists {
			utils.Error(c, utils.ErrUnauthorized)
			c.Abort()
			return
		}
		role, _ := value.(domain.UserRole)
		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}
		utils.Error(c, utils.ErrForbidden)
		c.Abort()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
