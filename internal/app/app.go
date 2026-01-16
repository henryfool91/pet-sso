package app

import (
	"log/slog"
	"time"

	"github.com/henryfool91/pet-sso/internal/services/auth"
	"github.com/henryfool91/pet-sso/internal/storage/sqlite"

	grpcapp "github.com/henryfool91/pet-sso/internal/app/grpc"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)

	if err != nil {
		panic(err)
	}
	authServ := auth.New(log, storage, storage, storage, tokenTTL)
	grpcApp := grpcapp.New(log, authServ, grpcPort)
	return &App{GRPCSrv: grpcApp}
}
