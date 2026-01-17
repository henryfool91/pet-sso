package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/henryfool91/pet-sso-protos/gen/go/sso/v1"
	"github.com/henryfool91/pet-sso/tests/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0
	appId      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, suite := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	respReg, err := suite.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := suite.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appId,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	assert.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(suite.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func TestRegister_DublicatedRegister(t *testing.T) {
	ctx, suite := suite.New(t)

	email := gofakeit.Email()
	password := randomFakePassword()

	respReg, err := suite.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = suite.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, suite := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password: value is required",
		},
		{
			name:        "Register with empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email: value is required",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email: value is required",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			respReg, err := suite.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			assert.Empty(t, respReg.GetUserId())
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

}

func TestLogin_FailCases(t *testing.T) {
	ctx, suite := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appId       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       appId,
			expectedErr: "password: value is required",
		},
		{
			name:        "Login with empty Email",
			email:       "",
			password:    randomFakePassword(),
			appId:       appId,
			expectedErr: "email: value is required",
		},
		{
			name:        "Login with Both Empty",
			email:       "",
			password:    "",
			appId:       appId,
			expectedErr: "email: value is required",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appId:       appId,
			expectedErr: "invalid credentials",
		},
		{
			name:        "Login without AppId",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appId:       emptyAppID,
			expectedErr: "app_id: value is required",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			respReg, err := suite.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appId,
			})
			require.Error(t, err)
			assert.Empty(t, respReg.GetToken())
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

}
