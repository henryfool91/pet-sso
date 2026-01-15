package jwt_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/henryfool91/pet-sso/internal/domain/models"
	jwtPkg "github.com/henryfool91/pet-sso/internal/lib/jwt"
)

func TestNewToken_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		user        models.User
		app         models.App
		duration    time.Duration
		expectError bool
		alterSecret bool // меняем секрет при парсинге, чтобы проверить invalid token
	}{
		{
			name:     "valid token",
			user:     models.User{ID: 123, Email: "test@example.com"}, // int64
			app:      models.App{ID: 456, Secret: "supersecretkey"},   // int32
			duration: 2 * time.Hour,
		},
		{
			name:        "empty app secret",
			user:        models.User{ID: 123, Email: "test@example.com"},
			app:         models.App{ID: 456, Secret: ""},
			duration:    30 * time.Minute,
			expectError: true,
		},
		{
			name:        "invalid secret on parse",
			user:        models.User{ID: 123, Email: "test@example.com"},
			app:         models.App{ID: 456, Secret: "supersecretkey"},
			duration:    1 * time.Hour,
			alterSecret: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := jwtPkg.NewToken(tt.user, tt.app, tt.duration)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tokenString == "" {
					t.Fatal("expected token string, got empty string")
				}
			}

			secretToUse := tt.app.Secret
			if tt.alterSecret {
				secretToUse = "wrongsecret"
			}

			parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					t.Fatalf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretToUse), nil
			})

			if tt.alterSecret {
				if err == nil {
					t.Fatal("expected error due to wrong secret, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("failed to parse token: %v", err)
			}

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			if !ok || !parsedToken.Valid {
				t.Fatal("token claims are invalid")
			}

			// Проверка uid (int64)
			if int64(claims["uid"].(float64)) != tt.user.ID {
				t.Errorf("expected uid %d, got %v", tt.user.ID, claims["uid"])
			}

			// Проверка email
			if claims["email"] != tt.user.Email {
				t.Errorf("expected email %s, got %v", tt.user.Email, claims["email"])
			}

			// Проверка app_id (int32)
			if int32(claims["app_id"].(float64)) != tt.app.ID {
				t.Errorf("expected app_id %d, got %v", tt.app.ID, claims["app_id"])
			}

			// Проверка exp
			exp := int64(claims["exp"].(float64))
			if time.Until(time.Unix(exp, 0)) < tt.duration-time.Minute || time.Until(time.Unix(exp, 0)) > tt.duration+time.Minute {
				t.Errorf("expected exp around %v, got %v", time.Now().Add(tt.duration), time.Unix(exp, 0))
			}
		})
	}
}
