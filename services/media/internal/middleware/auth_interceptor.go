package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/flick/backend/pkg/auth"
	"github.com/flick/backend/pkg/config"
	user_proto "github.com/flick/backend/services/user/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserClaimsKey contextKey = "userClaims"
)

// AuthInterceptor creates a gRPC unary interceptor for JWT authentication with user validation
func AuthInterceptor(cfg *config.Config, userClient user_proto.UserServiceClient) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fmt.Printf("[MEDIA AUTH] Processing gRPC request: %s\n", info.FullMethod)

		// Allow GetFile requests from other services without authentication
		if info.FullMethod == "/media.MediaService/GetFile" {
			fmt.Printf("[MEDIA AUTH] GetFile request - allowing without authentication for service-to-service calls\n")
			return handler(ctx, req)
		}

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

		// Verify user still exists in database
		if userClient != nil {
			_, err = userClient.GetUser(ctx, &user_proto.GetUserRequest{UserId: claims.UserID})
			if err != nil {
				fmt.Printf("[MEDIA AUTH] User validation failed: %v\n", err)
				return nil, status.Errorf(codes.Unauthenticated, "Your account is no longer active. Please log in again.")
			}
			fmt.Printf("[MEDIA AUTH] User existence validated for user: %s\n", claims.UserID)
		} else {
			fmt.Printf("[MEDIA AUTH] ERROR: User service client not available, rejecting request for security\n")
			return nil, status.Errorf(codes.Unavailable, "Service temporarily unavailable. Please try again later.")
		}

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
