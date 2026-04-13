package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"field-service/clients"
	"field-service/common/response"
	"field-service/config"
	"field-service/constants"
	errConstant "field-service/constants/error"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		timeStamp := time.Now()
		latency := timeStamp.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logrus.WithFields(logrus.Fields{
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"latency_ms": latency.Milliseconds(),
			"client_ip":  clientIP,
			"request_id": requestID,
		}).Info("request completed")
	}
}

func HandlePanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get("request_id")
				logrus.WithFields(logrus.Fields{
					"request_id": requestID,
				}).Errorf("Recovered from panic: %v", r)
				c.JSON(http.StatusInternalServerError, response.WithRequestId(response.Response{
					Status:  constants.Error,
					Message: errConstant.ErrInternalServerError.Error(),
				}, fmt.Sprintf("%v", requestID)))
				c.Abort()
			}
		}()
		c.Next()
	}
}

func RateLimiter(limit *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tollbooth.LimitByRequest(limit, c.Writer, c.Request)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"path": c.Request.URL.Path,
			}).Warn("Rate limit exceeded")

			c.Header("Retry-After", "60")
			response.HttpResponse(response.ParamHTTPResp{
				Code: http.StatusTooManyRequests,
				Err:  errConstant.ErrTooManyRequests,
				Gin:  c,
			})
			c.Abort()
			return
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

func contains(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func CheckRole(roles []string, client clients.IClientRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := client.GetUser().GetUserByToken(c.Request.Context())
		if err != nil {
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		if !contains(roles, user.Role) {
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}
		c.Next()
	}
}

func AuthenticateWithoutToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := validateAPIKey(c)
		if err != nil {
			responseUnauthorized(c, err.Error())
			return
		}
		c.Next()
	}
}

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		token := c.GetHeader(constants.Authorization)
		if token == "" {
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}
		err = validateAPIKey(c)
		if err != nil {
			responseUnauthorized(c, err.Error())
			return
		}
		tokenString := extractBearerToken(token)
		tokenUser := c.Request.WithContext(context.WithValue(c.Request.Context(), constants.Token, tokenString))
		c.Request = tokenUser
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigins := config.Config.AllowedOrigins
		origin := c.GetHeader("Origin")

		allowed := false
		if len(allowedOrigins) == 0 || config.Config.AppEnv == "local" {
			allowed = true
		} else {
			for _, o := range allowedOrigins {
				if o == origin {
					allowed = true
					break
				}
			}
		}

		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) == 0 {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-apikey, x-request-at")
		c.Header("Access-Control-Max-Age", "43200")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
