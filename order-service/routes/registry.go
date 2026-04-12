package routes

import (
	"github.com/gin-gonic/gin"
	"order-service/clients"
	controllers "order-service/controllers/http"
	routes "order-service/routes/order"
)

type IRouteRegistry interface {
	Serve()
}

func NewRouteRegistry(group *gin.RouterGroup, controller controllers.IControllerRegistry, client clients.IClientRegistry) IRouteRegistry {
	return &Registry{
		controlller: controller,
		client:      client,
		group:       group,
	}
}

type Registry struct {
	controlller controllers.IControllerRegistry
	client      clients.IClientRegistry
	group       *gin.RouterGroup
}

func (r *Registry) Serve() {
	r.orderRoute().Run()
}

func (r *Registry) orderRoute() routes.IOrderRoute {
	return routes.NewOrderRoute(r.group, r.client, r.controlller)
}
