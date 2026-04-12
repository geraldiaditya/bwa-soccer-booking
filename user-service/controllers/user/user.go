package controllers

import (
	"net/http"
	errWrap "user-service/common/error"
	"user-service/common/response"
	"user-service/domain/dto"
	"user-service/services"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service services.IServiceRegistry
}

type IUserController interface {
	Login(*gin.Context)
	Register(*gin.Context)
	Update(*gin.Context)
	GetUserLogin(*gin.Context)
	GetUserByUUID(*gin.Context)
}

func NewUserController(service services.IServiceRegistry) IUserController {
	return &UserController{service: service}
}

// Login godoc
// @Summary      User login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login credentials"
// @Success      200   {object}  response.Response{data=dto.UserResponse}
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Router       /auth/login [post]
func (u *UserController) Login(context *gin.Context) {
	request := &dto.LoginRequest{}

	if err := context.ShouldBindJSON(request); err != nil {
		context.Error(err)
		context.Abort()
		return
	}
	if err := errWrap.ValidateStruct(request); err != nil {
		context.Error(err)
		context.Abort()
		return
	}

	user, err := u.service.GetUser().Login(context.Request.Context(), request)
	if err != nil {
		context.Error(err)
		context.Abort()
		return
	}
	response.HttpResponse(
		response.ParamHTTPResp{
			Code:  http.StatusOK,
			Gin:   context,
			Data:  user.User,
			Token: &user.Token,
		},
	)
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Registration data"
// @Success      200   {object}  response.Response{data=dto.UserResponse}
// @Failure      400   {object}  response.Response
// @Failure      422   {object}  response.Response
// @Router       /auth/register [post]
func (u *UserController) Register(context *gin.Context) {
	request := &dto.RegisterRequest{}

	if err := context.ShouldBindJSON(request); err != nil {
		context.Error(err)
		context.Abort()
		return
	}
	if err := errWrap.ValidateStruct(request); err != nil {
		context.Error(err)
		context.Abort()
		return
	}

	user, err := u.service.GetUser().Register(
		context.Request.Context(), request,
	)
	if err != nil {
		context.Error(err)
		context.Abort()
		return
	}
	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Gin:  context,
			Data: user.User,
		},
	)
}

// Update godoc
// @Summary      Update user profile
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path      string            true  "User UUID"
// @Param        body  body      dto.UpdateRequest true  "Update data"
// @Success      200   {object}  response.Response{data=dto.UserResponse}
// @Failure      400   {object}  response.Response
// @Failure      401   {object}  response.Response
// @Router       /auth/{uuid} [put]
func (u *UserController) Update(context *gin.Context) {
	request := &dto.UpdateRequest{}
	uuid := context.Param("uuid")

	if err := context.ShouldBindJSON(request); err != nil {
		context.Error(err)
		context.Abort()
		return
	}
	if err := errWrap.ValidateStruct(request); err != nil {
		context.Error(err)
		context.Abort()
		return
	}

	user, err := u.service.GetUser().Update(
		context.Request.Context(), request, uuid,
	)
	if err != nil {
		context.Error(err)
		context.Abort()
		return
	}
	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Gin:  context,
			Data: user,
		},
	)
}

// GetUserLogin godoc
// @Summary      Get currently logged-in user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=dto.UserResponse}
// @Failure      401  {object}  response.Response
// @Router       /auth/user [get]
func (u *UserController) GetUserLogin(context *gin.Context) {
	user, err := u.service.GetUser().GetUserLogin(context.Request.Context())
	if err != nil {
		context.Error(err)
		context.Abort()
		return
	}

	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Data: user,
			Gin:  context,
		},
	)
}

// GetUserByUUID godoc
// @Summary      Get user by UUID
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path      string  true  "User UUID"
// @Success      200   {object}  response.Response{data=dto.UserResponse}
// @Failure      401   {object}  response.Response
// @Failure      404   {object}  response.Response
// @Router       /auth/{uuid} [get]
func (u *UserController) GetUserByUUID(context *gin.Context) {
	user, err := u.service.GetUser().GetUserByUUID(
		context.Request.Context(), context.Param("uuid"),
	)
	if err != nil {
		context.Error(err)
		context.Abort()
		return
	}

	response.HttpResponse(
		response.ParamHTTPResp{
			Code: http.StatusOK,
			Data: user,
			Gin:  context,
		},
	)
}
