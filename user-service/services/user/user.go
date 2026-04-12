package services

import (
	"context"
	"strings"
	"time"
	"user-service/config"
	"user-service/constants"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	"user-service/domain/models"
	"user-service/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repository repositories.IRepositoryRegistry
}

type IUserService interface {
	Login(context.Context, *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(context.Context, *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Update(context.Context, *dto.UpdateRequest, string) (*dto.UserResponse, error)
	GetUserLogin(context.Context) (*dto.UserResponse, error)
	GetUserByUUID(context.Context, string) (*dto.UserResponse, error)
}

type Claims struct {
	User *dto.UserResponse
	jwt.RegisteredClaims
}

func NewUserService(repository repositories.IRepositoryRegistry) IUserService {
	return &UserService{repository: repository}
}

func (u *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := u.repository.GetUser().FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errConstant.ErrWrongPassword
	}

	expTime := time.Now().Add(time.Duration(config.Config.JwtExpirationTime) * time.Minute).Unix()
	data := &dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		Role:        strings.ToLower(user.Role.Code),
	}

	claims := &Claims{
		User: data,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expTime, 0)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))

	if err != nil {
		return nil, err
	}
	response := &dto.LoginResponse{
		User:  *data,
		Token: tokenString,
	}
	return response, nil
}

// Helper methods removed to prevent N+1 queries

func (u *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	checkUsername, _ := u.repository.GetUser().FindByUsername(ctx, req.Username)
	if checkUsername != nil {
		return nil, errConstant.ErrUserAlreadyExist
	}

	checkEmail, _ := u.repository.GetUser().FindByEmail(ctx, req.Email)
	if checkEmail != nil {
		return nil, errConstant.ErrEmailAlreadyExist
	}

	if req.Password != req.ConfirmPassword {
		return nil, errConstant.ErrPasswordDoesNotMatch
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := u.repository.GetUser().Register(ctx, &dto.RegisterRequest{
		Name:        req.Name,
		Username:    req.Username,
		Password:    string(hashedPassword),
		Email:       req.Email,
		RoleID:      constants.Customer,
		PhoneNumber: req.PhoneNumber,
	})

	if err != nil {
		return nil, err
	}

	response := &dto.RegisterResponse{
		User: dto.UserResponse{
			UUID:        user.UUID,
			Name:        user.Name,
			Username:    user.Username,
			Email:       user.Email,
			Role:        strings.ToLower(user.Role.Code),
			PhoneNumber: user.PhoneNumber,
		},
	}
	return response, nil
}

func (u *UserService) Update(ctx context.Context, req *dto.UpdateRequest, uuid string) (*dto.UserResponse, error) {
	var (
		password                  string
		checkUsername, checkEmail *models.User
		hashedPassword            []byte
		user, userResult          *models.User
		err                       error
		data                      dto.UserResponse
	)

	user, err = u.repository.GetUser().FindByUUID(ctx, uuid)

	if err != nil {
		return nil, err
	}
	if user.Username != req.Username {
		checkUsername, _ = u.repository.GetUser().FindByUsername(ctx, req.Username)
		if checkUsername != nil {
			return nil, errConstant.ErrUserNameExist
		}
	}
	
	if user.Email != req.Email {
		checkEmail, _ = u.repository.GetUser().FindByEmail(ctx, req.Email)
		if checkEmail != nil {
			return nil, errConstant.ErrEmailAlreadyExist
		}
	}

	if req.Password != nil {
		if req.ConfirmPassword == nil || *req.Password != *req.ConfirmPassword {
			return nil, errConstant.ErrPasswordDoesNotMatch
		}
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		password = string(hashedPassword)
	}

	userResult, err = u.repository.GetUser().Update(ctx, &dto.UpdateRequest{
		Name:        req.Name,
		Username:    req.Username,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    &password,
	}, uuid)

	if err != nil {
		return nil, err
	}

	data = dto.UserResponse{
		UUID:        userResult.UUID,
		Name:        userResult.Name,
		Username:    userResult.Username,
		Email:       userResult.Email,
		PhoneNumber: userResult.PhoneNumber,
		Role:        strings.ToLower(userResult.Role.Code),
	}
	return &data, nil
}

func (u *UserService) GetUserLogin(ctx context.Context) (*dto.UserResponse, error) {
	userLogin, ok := ctx.Value(constants.UserLogin).(*dto.UserResponse)
	if !ok || userLogin == nil {
		return nil, errConstant.ErrUnauthorized
	}
	var (
		data dto.UserResponse
	)
	data = dto.UserResponse{
		UUID:        userLogin.UUID,
		Name:        userLogin.Name,
		Username:    userLogin.Username,
		Email:       userLogin.Email,
		Role:        userLogin.Role,
		PhoneNumber: userLogin.PhoneNumber,
	}
	return &data, nil
}

func (u *UserService) GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	user, err := u.repository.GetUser().FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	data := dto.UserResponse{
		UUID:        user.UUID,
		Name:        user.Name,
		Username:    user.Username,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        strings.ToLower(user.Role.Code),
	}
	return &data, nil
}
