package routes

import (
	"github.com/gin-gonic/gin"
	"payment-service/clients"
	controllers "payment-service/controllers/http"
	routes "payment-service/routes/payment"
)

type IRouteRegistry interface {
	Serve()
}
type Registry struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

func (r *Registry) Serve() {
	r.paymentRoute().Run()
}

func (r *Registry) paymentRoute() routes.IPaymentRoute {
	return routes.NewPaymentRoute(r.controller, r.client, r.group)
}

func NewRouteRegistry(
	group *gin.RouterGroup,
	controller controllers.IControllerRegistry,
	client clients.IClientRegistry,
) IRouteRegistry {
	return &Registry{
		controller: controller,
		group:      group,
		client:     client,
	}
}
