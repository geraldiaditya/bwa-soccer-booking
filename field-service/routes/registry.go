package routes

import (
	"field-service/clients"
	"field-service/controllers"
	rField "field-service/routes/field"
	rFieldSchedule "field-service/routes/field_schedule"
	rTime "field-service/routes/time"
	"github.com/gin-gonic/gin"
)

type IRegistry interface {
	Serve()
}

func NewRouteRegistry(group *gin.RouterGroup, controller controllers.IControllerRegistry, client clients.IClientRegistry) IRegistry {
	return &Registry{
		controller: controller,
		group:      group,
		client:     client,
	}
}

type Registry struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
	client     clients.IClientRegistry
}

func (r *Registry) fieldRoute() rField.IFieldRoute {
	return rField.NewFieldRoute(r.group, r.controller, r.client)
}

func (r *Registry) fieldScheduleRoute() rFieldSchedule.IFieldScheduleRoute {
	return rFieldSchedule.NewFieldScheduleRoute(r.group, r.controller, r.client)
}

func (r *Registry) timeRoute() rTime.ITimeRoute {
	return rTime.NewTimeRoute(r.group, r.controller, r.client)
}

func (r *Registry) Serve() {
	r.fieldRoute().Run()
	r.fieldScheduleRoute().Run()
	r.timeRoute().Run()
}
