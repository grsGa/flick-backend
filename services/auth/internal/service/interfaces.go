package service

import (
	"context"

	"github.com/flick/backend/services/auth/proto"
)

// AuthService defines the interface for the authentication service.
type AuthService interface {
	Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error)
	Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error)
	ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error)
	RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error)
	Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error)
	GithubLogin(ctx context.Context, req *proto.GithubLoginRequest) (*proto.GithubLoginResponse, error)
	GithubCallback(ctx context.Context, req *proto.GithubCallbackRequest) (*proto.GithubCallbackResponse, error)
	GoogleLogin(ctx context.Context, req *proto.GoogleLoginRequest) (*proto.GoogleLoginResponse, error)
	GoogleCallback(ctx context.Context, req *proto.GoogleCallbackRequest) (*proto.GoogleCallbackResponse, error)
}
