package routes

import (
	"github.com/gin-gonic/gin"
	"user-service/controllers"
	routes "user-service/routes/user"
)

type Registry struct {
	controller controllers.IControllerRegistry
	group      *gin.RouterGroup
}

type IRouterRegistry interface {
	Serve()
}

func NewRouteRegistry(controller controllers.IControllerRegistry, group *gin.RouterGroup) IRouterRegistry {
	return &Registry{controller, group}
}
func (r *Registry) userRoute() routes.IUserRoute {
	return routes.NewUserRoute(r.controller, r.group)
}
func (r *Registry) Serve() {
	r.userRoute().Run()
}
