package routes

import (
	"field-service/clients"
	"field-service/constants"
	"field-service/controllers"
	"field-service/middlewares"
	"github.com/gin-gonic/gin"
)

type ITimeRoute interface {
	Run()
}

func NewTimeRoute(group *gin.RouterGroup, controller controllers.IControllerRegistry, client clients.IClientRegistry) ITimeRoute {
	return &TimeRoute{
		controller: controller,
		group:      group,
		client:     client,
	}
}

type TimeRoute struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

func (t *TimeRoute) Run() {
	group := t.group.Group("/time")
	group.Use(middlewares.Authenticate())
	group.GET("", middlewares.CheckRole([]string{
		constants.Admin,
	}, t.client), t.controller.GetTime().GetAll)
	group.GET("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
	}, t.client), t.controller.GetTime().GetByUUID)
	group.POST("", middlewares.CheckRole([]string{
		constants.Admin,
	}, t.client), t.controller.GetTime().Create)
}
