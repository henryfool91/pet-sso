package auth

import (
	"context"

	ssov1 "github.com/henryfool91/pet-sso-protos/gen/go/sso"
	"google.golang.org/grpc"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	panic("not implemented")
}

// IsAdmin implements [ssov1.AuthServer].
func (s *serverAPI) IsAdmin(context.Context, *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("unimplemented")
}

// Register implements [ssov1.AuthServer].
func (s *serverAPI) Register(context.Context, *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	panic("unimplemented")
}
