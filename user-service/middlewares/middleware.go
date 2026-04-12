package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	errWrap "user-service/common/error"
	"user-service/common/response"
	"user-service/config"
	"user-service/constants"
	errConstant "user-service/constants/error"
	"user-service/domain/dto"
	services "user-service/services/user"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

func HandlePanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Recovered from panic: %v", r)
				c.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,
					Message: errConstant.ErrInternalServerError.Error(),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var fieldErrors validator.ValidationErrors
			
			if errors.As(err, &fieldErrors) {
				errMessage := http.StatusText(http.StatusUnprocessableEntity)
				errResponse := errWrap.ErrValidationResponse(err)
				response.HttpResponse(response.ParamHTTPResp{
					Code:    http.StatusUnprocessableEntity,
					Message: &errMessage,
					Data:    errResponse,
					Err:     err,
					Gin:     c,
				})
				return
			}
			
			code := http.StatusBadRequest
			if errConstant.ErrMapping(err) {
				switch {
				case errors.Is(err, errConstant.ErrUnauthorized):
					code = http.StatusUnauthorized
				case errors.Is(err, errConstant.ErrForbidden):
					code = http.StatusForbidden
				case errors.Is(err, errConstant.ErrNotFound), errors.Is(err, errConstant.ErrUserNotFound):
					code = http.StatusNotFound
				case errors.Is(err, errConstant.ErrInternalServerError), errors.Is(err, errConstant.ErrSQLError):
					code = http.StatusInternalServerError
				}
			}

			response.HttpResponse(response.ParamHTTPResp{
				Code: code,
				Err:  err,
				Gin:  c,
			})
		}
	}
}

func RateLimiter(limit *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tollbooth.LimitByRequest(limit, c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusTooManyRequests, response.Response{
				Status:  constants.Error,
				Message: errConstant.ErrTooManyRequests.Error(),
			})
			c.Abort()
		}
		c.Next()
	}
}

func extractBearerToken(token string) string {
	arrayToken := strings.Split(token, " ")
	if len(arrayToken) == 2 {
		return arrayToken[1]
	}
	return ""
}

func responseUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error,
		Message: message,
	})
	c.Abort()
}

func validateAPIKey(c *gin.Context) error {
	apiKey := c.GetHeader(constants.XApiKey)
	requestAt := c.GetHeader(constants.XRequestAt)
	serviceName := c.GetHeader(constants.XServiceName)

	signatureKey := config.Config.SignatureKey

	validateKey := fmt.Sprintf("%s:%s:%s", serviceName, signatureKey, requestAt)

	hash := sha256.New()
	hash.Write([]byte(validateKey))
	resultHash := hex.EncodeToString(hash.Sum(nil))

	if apiKey != resultHash {
		return errConstant.ErrUnauthorized
	}
	return nil
}

func validateBearerToken(c *gin.Context, token string) error {
	if !strings.Contains(token, "Bearer") {
		return errConstant.ErrUnauthorized
	}

	tokenString := extractBearerToken(token)
	if tokenString == "" {
		return errConstant.ErrUnauthorized
	}

	claims := &services.Claims{}
	tokenJwt, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errConstant.ErrWrongPassword
		}
		jwtSecret := []byte(config.Config.JwtSecretKey)
		return jwtSecret, nil
	})
	if err != nil || !tokenJwt.Valid {
		return errConstant.ErrUnauthorized
	}
	userLogin := c.Request.WithContext(context.WithValue(c.Request.Context(), constants.UserLogin, claims.User))
	c.Request = userLogin
	c.Set(constants.Token, token)
	return nil
}

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		token := c.GetHeader(constants.Authorization)
		if token == "" {
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}
		err = validateBearerToken(c, token)
		if err != nil {
			responseUnauthorized(c, err.Error())
			return
		}
		err = validateAPIKey(c)
		if err != nil {
			responseUnauthorized(c, err.Error())
			return
		}
		c.Next()
	}
}

// AuthorizeOwnerOrAdmin ensures the authenticated user can only access their own
// resource (matched by :uuid path param) unless they are an admin.
func AuthorizeOwnerOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userLogin, ok := c.Request.Context().Value(constants.UserLogin).(*dto.UserResponse)
		if !ok || userLogin == nil {
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}
		paramUUID := c.Param("uuid")
		isOwner := userLogin.UUID.String() == paramUUID
		isAdmin := userLogin.Role == "admin"
		if !isOwner && !isAdmin {
			c.JSON(http.StatusForbidden, response.Response{
				Status:  constants.Error,
				Message: errConstant.ErrForbidden.Error(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
