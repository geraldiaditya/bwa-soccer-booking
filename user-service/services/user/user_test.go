package services

import (
	"context"
	"testing"
	"user-service/config"
	"user-service/constants"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	"user-service/domain/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	config.Config.JwtSecretKey = "test-secret"
	config.Config.JwtExpirationTime = 60
}

func newService() (IUserService, *mockRepositoryRegistry) {
	reg := newMockRegistry()
	return NewUserService(reg), reg
}

// ── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	svc, reg := newService()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	user := &models.User{
		UUID:     uuid.New(),
		Username: "john",
		Password: string(hashed),
		Email:    "john@example.com",
		Role:     models.Role{Code: "CUSTOMER"},
	}

	reg.userRepo.On("FindByUsername", context.Background(), "john").Return(user, nil)

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "john",
		Password: "secret",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "john", resp.User.Username)
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, reg := newService()

	reg.userRepo.On("FindByUsername", context.Background(), "ghost").
		Return(nil, errConstant.ErrUserNotFound)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "ghost",
		Password: "any",
	})

	assert.ErrorIs(t, err, errConstant.ErrUserNotFound)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, reg := newService()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	user := &models.User{Username: "john", Password: string(hashed)}

	reg.userRepo.On("FindByUsername", context.Background(), "john").Return(user, nil)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Username: "john",
		Password: "wrong",
	})

	assert.ErrorIs(t, err, errConstant.ErrWrongPassword)
}

// ── Register ─────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	svc, reg := newService()

	reg.userRepo.On("FindByUsername", context.Background(), "jane").
		Return(nil, errConstant.ErrUserNotFound)
	reg.userRepo.On("FindByEmail", context.Background(), "jane@example.com").
		Return(nil, errConstant.ErrUserNotFound)

	createdUser := &models.User{
		UUID:     uuid.New(),
		Name:     "Jane",
		Username: "jane",
		Email:    "jane@example.com",
		RoleID:   constants.Customer,
		Role:     models.Role{Code: "CUSTOMER"},
	}
	reg.userRepo.On("Register", context.Background(), mock_registerReq("jane")).
		Return(createdUser, nil)

	resp, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Name:            "Jane",
		Username:        "jane",
		Password:        "pass1234",
		ConfirmPassword: "pass1234",
		Email:           "jane@example.com",
		PhoneNumber:     "08123456789",
	})

	assert.NoError(t, err)
	assert.Equal(t, "jane", resp.User.Username)
}

func TestRegister_UsernameAlreadyExists(t *testing.T) {
	svc, reg := newService()

	reg.userRepo.On("FindByUsername", context.Background(), "taken").
		Return(&models.User{Username: "taken"}, nil)

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username:        "taken",
		Password:        "pass1234",
		ConfirmPassword: "pass1234",
		Email:           "new@example.com",
	})

	assert.ErrorIs(t, err, errConstant.ErrUserAlreadyExist)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	svc, reg := newService()

	reg.userRepo.On("FindByUsername", context.Background(), "newuser").
		Return(nil, errConstant.ErrUserNotFound)
	reg.userRepo.On("FindByEmail", context.Background(), "taken@example.com").
		Return(&models.User{Email: "taken@example.com"}, nil)

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username:        "newuser",
		Password:        "pass1234",
		ConfirmPassword: "pass1234",
		Email:           "taken@example.com",
	})

	assert.ErrorIs(t, err, errConstant.ErrEmailAlreadyExist)
}

func TestRegister_PasswordMismatch(t *testing.T) {
	svc, reg := newService()

	reg.userRepo.On("FindByUsername", context.Background(), "alice").
		Return(nil, errConstant.ErrUserNotFound)
	reg.userRepo.On("FindByEmail", context.Background(), "alice@example.com").
		Return(nil, errConstant.ErrUserNotFound)

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Username:        "alice",
		Password:        "pass1234",
		ConfirmPassword: "different",
		Email:           "alice@example.com",
	})

	assert.ErrorIs(t, err, errConstant.ErrPasswordDoesNotMatch)
}

// ── GetUserLogin ─────────────────────────────────────────────────────────────

func TestGetUserLogin_Success(t *testing.T) {
	svc, _ := newService()

	userResp := &dto.UserResponse{
		UUID:     uuid.New(),
		Username: "john",
		Email:    "john@example.com",
		Role:     "customer",
	}
	ctx := context.WithValue(context.Background(), constants.UserLogin, userResp)

	resp, err := svc.GetUserLogin(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "john", resp.Username)
}

func TestGetUserLogin_MissingContext(t *testing.T) {
	svc, _ := newService()

	_, err := svc.GetUserLogin(context.Background())

	assert.ErrorIs(t, err, errConstant.ErrUnauthorized)
}

// ── GetUserByUUID ─────────────────────────────────────────────────────────────

func TestGetUserByUUID_Success(t *testing.T) {
	svc, reg := newService()
	id := uuid.New()

	reg.userRepo.On("FindByUUID", context.Background(), id.String()).
		Return(&models.User{UUID: id, Username: "john", Email: "john@example.com"}, nil)

	resp, err := svc.GetUserByUUID(context.Background(), id.String())

	assert.NoError(t, err)
	assert.Equal(t, id, resp.UUID)
}

func TestGetUserByUUID_NotFound(t *testing.T) {
	svc, reg := newService()
	id := uuid.New()

	reg.userRepo.On("FindByUUID", context.Background(), id.String()).
		Return(nil, errConstant.ErrUserNotFound)

	_, err := svc.GetUserByUUID(context.Background(), id.String())

	assert.ErrorIs(t, err, errConstant.ErrUserNotFound)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// mock_registerReq returns a testify matcher that checks only the non-hashed fields
// (password will be hashed so we can't match it exactly).
func mock_registerReq(username string) interface{} {
	return mock.MatchedBy(func(req *dto.RegisterRequest) bool {
		return req.Username == username
	})
}
