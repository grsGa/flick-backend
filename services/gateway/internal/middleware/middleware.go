package middleware

import (
	"context"
	"strings"

	"backend/pkg/auth"
	"backend/pkg/config"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	UserClaimsKey contextKey = "userClaims"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		claims, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			c.Next()
			return
		}

		ctx := context.WithValue(c.Request.Context(), UserClaimsKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetUserClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
