package routes

import (
	"github.com/gin-gonic/gin"
	"order-service/clients"
	"order-service/constants"
	controllers "order-service/controllers/http"
	"order-service/middlewares"
)

func NewOrderRoute(group *gin.RouterGroup, client clients.IClientRegistry, controller controllers.IControllerRegistry) IOrderRoute {
	return &OrderRoute{
		controlller: controller,
		client:      client,
		group:       group,
	}
}

type OrderRoute struct {
	controlller controllers.IControllerRegistry
	client      clients.IClientRegistry
	group       *gin.RouterGroup
}

func (o *OrderRoute) Run() {
	group := o.group.Group("order")
	group.Use(middlewares.Authenticate())
	group.GET("", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, o.client), o.controlller.GetOrder().GetAllWithPagination)
	group.GET("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, o.client), o.controlller.GetOrder().GetByUUID)
	group.GET("/user", middlewares.CheckRole([]string{
		constants.Customer,
	}, o.client), o.controlller.GetOrder().GetOrderByUserId)

	group.POST("", middlewares.CheckRole([]string{
		constants.Admin,
	}, o.client), o.controlller.GetOrder().Create)

}

type IOrderRoute interface {
	Run()
}
