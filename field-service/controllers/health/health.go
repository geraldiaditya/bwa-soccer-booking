package health

import (
	"field-service/common/response"
	"field-service/constants"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IHealthController interface {
	Health(c *gin.Context)
	Ready(c *gin.Context)
}

type HealthController struct {
	db *gorm.DB
}

func NewHealthController(db *gorm.DB) IHealthController {
	return &HealthController{db: db}
}

func (h *HealthController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, response.Response{
		Status:  constants.Success,
		Message: "healthy",
	})
}

func (h *HealthController) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, response.Response{
			Status:  constants.Error,
			Message: "database connection error",
		})
		return
	}

	err = sqlDB.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, response.Response{
			Status:  constants.Error,
			Message: "database ping error",
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Status:  constants.Success,
		Message: "ready",
	})
}
