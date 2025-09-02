package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/flick/backend/pkg/auth"
	"github.com/flick/backend/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserClaimsKey contextKey = "userClaims"
)

// AuthInterceptor creates a gRPC unary interceptor for JWT authentication
func AuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fmt.Printf("[MEDIA AUTH] Processing gRPC request: %s\n", info.FullMethod)

		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			fmt.Printf("[MEDIA AUTH] No metadata found in context\n")
			return nil, status.Errorf(codes.Unauthenticated, "no metadata found")
		}

		// Get authorization header from metadata
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			fmt.Printf("[MEDIA AUTH] No authorization header found\n")
			return nil, status.Errorf(codes.Unauthenticated, "no authorization header")
		}

		authHeader := authHeaders[0]
		fmt.Printf("[MEDIA AUTH] Authorization header found: %s...\n", authHeader[:min(len(authHeader), 20)])

		// Extract Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			fmt.Printf("[MEDIA AUTH] Invalid Bearer token format\n")
			return nil, status.Errorf(codes.Unauthenticated, "invalid token format")
		}

		// Validate JWT token
		claims, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
		if err != nil {
			fmt.Printf("[MEDIA AUTH] JWT validation failed: %v\n", err)
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		fmt.Printf("[MEDIA AUTH] JWT validated successfully for user: %s\n", claims.UserID)

		// Add claims to context
		ctx = context.WithValue(ctx, UserClaimsKey, claims)

		// Call the handler with authenticated context
		return handler(ctx, req)
	}
}

// GetUserClaims extracts user claims from context
func GetUserClaims(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
