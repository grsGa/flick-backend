package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/flick/backend/pkg/auth"
	"github.com/flick/backend/pkg/config"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	UserClaimsKey contextKey = "userClaims"
	TokenKey      contextKey = "token"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Debug logging
		fmt.Printf("[AUTH] Processing request: %s %s\n", c.Request.Method, c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Printf("[AUTH] No Authorization header found\n")
			c.Next()
			return
		}

		fmt.Printf("[AUTH] Authorization header found: %s...\n", authHeader[:min(len(authHeader), 20)])

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			fmt.Printf("[AUTH] Invalid Bearer token format\n")
			c.Next()
			return
		}

		claims, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			fmt.Printf("[AUTH] JWT validation failed: %v\n", err)
			c.Next()
			return
		}

		fmt.Printf("[AUTH] JWT validated successfully for user: %s\n", claims.UserID)
		ctx := context.WithValue(c.Request.Context(), UserClaimsKey, claims)
		ctx = context.WithValue(ctx, TokenKey, tokenString)
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

func GetTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(TokenKey).(string); ok {
		return token
	}
	return ""
}
