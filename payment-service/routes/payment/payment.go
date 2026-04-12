package routes

import (
	"github.com/gin-gonic/gin"
	"payment-service/clients"
	"payment-service/constants"
	controllers "payment-service/controllers/http"
	"payment-service/middlewares"
)

type IPaymentRoute interface {
	Run()
}
type PaymentRoute struct {
	controller controllers.IControllerRegistry
	client     clients.IClientRegistry
	group      *gin.RouterGroup
}

func (p *PaymentRoute) Run() {
	group := p.group.Group("/payment")
	group.POST("/webhook", p.controller.GetPayment().Webhook)

	group.Use(middlewares.Authenticate())

	group.GET("", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, p.client), p.controller.GetPayment().GetAllWithPagination)

	group.GET("/:uuid", middlewares.CheckRole([]string{
		constants.Admin,
		constants.Customer,
	}, p.client), p.controller.GetPayment().GetByUUID)

	group.POST("", middlewares.CheckRole([]string{
		constants.Customer,
	}, p.client), p.controller.GetPayment().Create)
}

func NewPaymentRoute(
	controller controllers.IControllerRegistry,
	client clients.IClientRegistry,
	group *gin.RouterGroup,
) IPaymentRoute {
	return &PaymentRoute{
		controller: controller,
		client:     client,
		group:      group,
	}
}
