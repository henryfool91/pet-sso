package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/henryfool91/pet-sso/internal/domain/models"
	"github.com/henryfool91/pet-sso/internal/lib/jwt"
	"github.com/henryfool91/pet-sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidAppId       = errors.New("invalid app id")
)

// New returns a new instance of the Auth service
func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int32) (token string, err error) {
	op := "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("logging user")
	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFoud) {
			a.log.Warn("user not found", slog.String("err", err.Error()))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)

		}
		log.Error("failed to get user", slog.String("err", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", slog.String("err", err.Error()))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)

	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found")
			return "", fmt.Errorf("%s: %w", op, ErrInvalidAppId)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged successfully")
	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", slog.String("err", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

func (a *Auth) Register(ctx context.Context, email string, password string) (userID int64, err error) {
	op := "auth.Register"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password hash", slog.String("err", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists")
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", slog.String("err", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error) {
	op := "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("userID", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err = a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("checked if user is admin")
	return isAdmin, nil
}
