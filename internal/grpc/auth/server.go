package auth

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	ssov1 "github.com/henryfool91/pet-sso-protos/gen/go/sso/v1"
	"github.com/henryfool91/pet-sso/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
		appID int32) (token string, err error)

	Register(ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context,
		userID int64) (isAdmin bool, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

// Login implements [ssov1.AuthServer].
func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		} else if errors.Is(err, auth.ErrInvalidAppId) {
			return nil, status.Error(codes.InvalidArgument, "invalid app id")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.LoginResponse{Token: token}, nil
}

// Register implements [ssov1.AuthServer].
func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.RegisterResponse{UserId: userID}, nil
}

// IsAdmin implements [ssov1.AuthServer].
func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}
