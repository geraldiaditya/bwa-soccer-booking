package health

import (
	healthController "field-service/controllers/health"
	"github.com/gin-gonic/gin"
)

type HealthRoute struct {
	router     *gin.Engine
	controller healthController.IHealthController
}

func NewHealthRoute(router *gin.Engine, controller healthController.IHealthController) *HealthRoute {
	return &HealthRoute{
		router:     router,
		controller: controller,
	}
}

func (r *HealthRoute) Serve() {
	r.router.GET("/health", r.controller.Health)
	r.router.GET("/ready", r.controller.Ready)
}
